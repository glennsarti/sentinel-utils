var http = require('https');
var fs = require("fs");
var path = require("path");

console.log(`Running ${__filename}`)

const goplsRef = "gopls/v0.11.0";
const goplsRootUrl = "https://raw.githubusercontent.com/golang/tools/" + goplsRef + "/gopls/internal/lsp/protocol/";

const scriptDir = __dirname;
const root = path.join(scriptDir, "..");
const outDir = path.join(root, "lib","languageserver","internal","protocol");

function cleanProtocolFiles(outDir) {
  console.log(`Cleaning LSP protocol files from ${outDir} ...`);

  const files = fs.readdirSync(outDir);
  const goFiles = files.filter(file =>
    path.extname(file) === '.go' && file.startsWith('generated_')
  );

  goFiles.forEach(file => {
    const filePath = path.join(outDir, file);
    console.log(`Deleted: ${file}`);
    fs.unlinkSync(filePath);
  });
}

function getUrlAsPromise(url) {
  return new Promise((resolve, reject) => {
    console.log(`Attempting to download ${url} ...`);
    http.get(url, (response) => {
      let chunksOfData = [];

      response.on('data', (fragments) => {
        chunksOfData.push(fragments);
      });

      response.on('end', () => {
        let responseBody = Buffer.concat(chunksOfData);
        resolve({
          resultCode: response.statusCode,
          redirect: response.headers.location,
          body: responseBody.toString()
        });
      });

      response.on('error', (error) => {
        reject(error);
      });
    });
  });
};

async function getProtocolFiles() {
  let content = await getUrlAsPromise(goplsRootUrl + "tsdocument_changes.go");
  let outFile = path.join(outDir, "generated_protocol_document_changes.go");
  console.log(`Writing ${outFile}...`);
  fs.writeFileSync(outFile, content.body, { encoding: "utf8" });

  content = await getUrlAsPromise(goplsRootUrl + "tsprotocol.go");
  outFile = path.join(outDir, "generated_protocol.go");
  console.log(`Writing ${outFile}...`);
  fs.writeFileSync(outFile, content.body, { encoding: "utf8" });
};

(async function () {
  console.log(`\nDownloading LSP protocol files for ${goplsRef}\n`);

  cleanProtocolFiles(outDir);
  await getProtocolFiles();

  process.exit(0);
})();
