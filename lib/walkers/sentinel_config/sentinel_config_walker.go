package walkers

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/glennsarti/sentinel-parser/features"
	"github.com/glennsarti/sentinel-parser/filetypes"
	"github.com/glennsarti/sentinel-utils/lib/filesystem"
	"github.com/glennsarti/sentinel-utils/lib/internal/helpers"
	"github.com/glennsarti/sentinel-utils/lib/parsing"

	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

type Visitor func(file *filesystem.File) (bool, error)

type Walker interface {
	Walk(visitor Visitor) error
	SentinelVersion() string
	FileSystem() filesystem.FS
	Root() string
}

// defaultConfigHCL is the default Sentinel configuration HCL file.
const defaultConfigHCL = `sentinel.hcl`

// defaultConfigJSON is the default Sentinel configuration JSON file.
const defaultConfigJSON = `sentinel.json`

type sentinelConfigWalker struct {
	root            string
	sentinelVersion string
	fsys            filesystem.FS
	parsing         parsing.Factory
}

func NewSentinelConfigWalker(fsys filesystem.FS, root, sentinelVersion string, pf parsing.Factory) Walker {
	walker := sentinelConfigWalker{
		root:            root,
		fsys:            fsys,
		sentinelVersion: sentinelVersion,
		parsing:         pf,
	}

	return &walker
}

// TODO: Not sure I need this yet?
func nodeDocumentID(node ast.HCLNode) string {
	return fmt.Sprintf("%s-%s", node.BlockType(), node.BlockName())
}

func (dw *sentinelConfigWalker) SentinelVersion() string {
	return dw.sentinelVersion
}

func (dw *sentinelConfigWalker) FileSystem() filesystem.FS {
	return dw.fsys
}

func (dw *sentinelConfigWalker) Root() string {
	return dw.root
}

func (dw *sentinelConfigWalker) Walk(visitor Visitor) error {
	cfgPath, cfgName, err := dw.getRootConfig()
	if err != nil {
		return err
	}
	rootFile := &filesystem.File{
		Path: cfgPath,
		Name: cfgName,
		Type: filetypes.ConfigPrimaryFileType,
	}

	// Order is important.
	// First visit the root file
	if cont, err := visitor(rootFile); err != nil || !cont {
		return err
	}

	if features.SupportedVersion(dw.sentinelVersion, features.ConfigurationOverrideMinimumVersion) {
		// Then visit the overrides
		if err := dw.visitOverrideFiles(rootFile, visitor); err != nil {
			return err
		}
	}

	// Then visit items defined in the root file
	return dw.recurseRootConfig(rootFile, dw.sentinelVersion, visitor)
}

// Returns the path and filename of the root configuration file
func (dw *sentinelConfigWalker) getRootConfig() (string, string, error) {
	i, err := fs.Stat(dw.fsys, dw.root)
	if err != nil {
		return "", "", err
	}
	de := fs.FileInfoToDirEntry(i)

	// It's a file so use that.
	if !de.IsDir() {
		return dw.root, de.Name(), nil
	}

	// Check for the default HCL config file
	filename := dw.fsys.PathJoin(dw.root, defaultConfigHCL)
	if _, err := fs.Stat(dw.fsys, filename); err == nil {
		return filename, defaultConfigHCL, nil
	}

	// Check for the default JSON config file
	filename = dw.fsys.PathJoin(dw.root, defaultConfigJSON)
	if _, err := fs.Stat(dw.fsys, filename); err == nil {
		return "", "", fmt.Errorf("the sentinel configuration file %q is not supported", filename)
		//return filename, defaultConfigJSON, nil
	}

	return "", "", fmt.Errorf("could not find a Sentinel configuration file in directory %s", dw.root)
}

func (dw *sentinelConfigWalker) isOverride(item fs.DirEntry) bool {
	return item.Name() == "override.hcl" ||
		strings.HasSuffix(item.Name(), "_override.hcl") ||
		item.Name() == "override.json" ||
		strings.HasSuffix(item.Name(), "_override.json")
}

func (dw *sentinelConfigWalker) visitOverrideFiles(rootFile *filesystem.File, visitor Visitor) error {
	rootExt := ""
	if strings.HasSuffix(rootFile.Name, ".hcl") {
		rootExt = ".hcl"
	}
	if strings.HasSuffix(rootFile.Name, ".json") {
		rootExt = ".json"
	}
	if rootExt == "" {
		return fmt.Errorf("%s does not have a valid extension for a root configuration file", rootFile.Name)
	}

	items, err := dw.fsys.ReadDir(dw.root)
	if err != nil {
		return err
	}
	for _, item := range items {
		// We only care about files that share the same extension as the root config file,
		// and have the correct filename
		if item.IsDir() ||
			!strings.HasSuffix(item.Name(), rootExt) ||
			!dw.isOverride(item) {
			continue
		}

		override := &filesystem.File{
			Path: dw.fsys.PathJoin(dw.root, item.Name()),
			Name: item.Name(),
			Type: filetypes.ConfigOverrideFileType,
		}
		if cont, err := visitor(override); err != nil || !cont {
			return err
		}
	}
	return nil
}

func (dw *sentinelConfigWalker) recurseRootConfig(rootFile *filesystem.File, sentinelVersion string, visitor Visitor) error {
	cfg, diags, err := dw.parsing.ParseSentinelConfigFile(rootFile, sentinelVersion)
	if err != nil {
		return err
	}
	if diags.HasErrors() {
		return diags
	}
	parentDir := dw.fsys.ParentPath(rootFile.Path)

	// Order is important here
	// Modules first
	// Policies
	//   Policy Tests

	keys := helpers.SortedKeys(cfg.Imports)
	// Figure out local modules
	for _, key := range keys {
		imp := cfg.Imports[key]
		if imp == nil {
			continue
		}

		modSource := ""
		switch actual := imp.(type) {
		case *ast.V1ModuleImport:
			modSource = actual.Source
		case *ast.V2ModuleImport:
			modSource = actual.Source
		}
		if strings.HasPrefix(modSource, "./") {
			_, cont, err := dw.visitFilePath(&filesystem.File{
				Path: dw.fsys.PathJoin(parentDir, modSource[2:]),
				Type: filetypes.ModuleFileType,
				ID:   nodeDocumentID(imp),
			}, visitor)

			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			if !cont {
				return nil
			}
		}
	}

	keys = helpers.SortedKeys(cfg.Policies)
	// Figure out all the policies
	for _, key := range keys {
		pol := cfg.Policies[key]
		if pol == nil {
			continue
		}
		if strings.HasPrefix(pol.Source, "./") {
			policyPath := dw.fsys.PathJoin(parentDir, pol.Source[2:])

			policyFile, cont, err := dw.visitFilePath(&filesystem.File{
				Path: policyPath,
				Type: filetypes.PolicyFileType,
				ID:   nodeDocumentID(pol),
			}, visitor)
			if err != nil {
				return err
			}
			if !cont {
				return nil
			}

			if err := dw.findPolicyTests(pol.Name, policyFile, visitor); err != nil {
				return err
			}
		}
	}

	return nil
}

func (dw *sentinelConfigWalker) visitFilePath(file *filesystem.File, visitor Visitor) (*filesystem.File, bool, error) {
	if i, err := fs.Stat(dw.fsys, file.Path); err != nil {
		return file, false, err
	} else {
		file.Name = fs.FileInfoToDirEntry(i).Name()
		cont, err := visitor(file)
		return file, cont, err
	}
}

func (dw *sentinelConfigWalker) findPolicyTests(policyName string, policyFile *filesystem.File, visitor Visitor) error {
	// Get the parent dir of the policyPath
	parent := dw.fsys.ParentPath(policyFile.Path)
	// See if <parent>/test/<policy name>/ dir exists
	testPath := dw.fsys.PathJoin(parent, "test", policyName)
	if _, err := fs.Stat(dw.fsys, testPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		} else {
			return err
		}
	}

	if entries, err := fs.ReadDir(dw.fsys, testPath); err != nil {
		return err
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if strings.HasSuffix(entry.Name(), ".hcl") || strings.HasSuffix(entry.Name(), ".json") {
				testFilePath := dw.fsys.PathJoin(testPath, entry.Name())
				if cont, err := visitor(&filesystem.File{
					Path: testFilePath,
					Name: entry.Name(),
					Type: filetypes.ConfigTestFileType,
				}); err != nil {
					return err
				} else if !cont {
					return nil
				}
			}
		}
	}
	return nil
}
