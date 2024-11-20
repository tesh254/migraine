const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const version = process.env.npm_package_version;
const platform = process.platform;
const arch = process.arch === "x64" ? "amd64" : process.arch;

const getBinaryName = () => {
  const os = platform === "darwin" ? "darwin" : "linux";
  return `migraine-${os}-${arch}`;
};

const getDownloadUrl = () => {
  const binaryName = getBinaryName();
  return `https://github.com/tesh254/migraine/releases/download/v${version}/${binaryName}`;
};

const download = (url, dest) => {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, (response) => {
        response.pipe(file);
        file.on("finish", () => {
          file.close();
          resolve();
        });
      })
      .on("error", (err) => {
        fs.unlink(dest, () => reject(err));
      });
  });
};

async function install() {
  try {
    const binPath = path.join(__dirname, "..", "bin");
    const binaryPath = path.join(binPath, "migraine");
    const aliasPath = path.join(binPath, "mig");

    // Create bin directory if it doesn't exist
    if (!fs.existsSync(binPath)) {
      fs.mkdirSync(binPath, { recursive: true });
    }

    // Download the binary
    const downloadUrl = getDownloadUrl();
    console.log(`Downloading migraine from ${downloadUrl}`);
    await download(downloadUrl, binaryPath);

    // Make binary executable
    fs.chmodSync(binaryPath, "755");

    // Create symbolic link for alias
    if (fs.existsSync(aliasPath)) {
      fs.unlinkSync(aliasPath);
    }
    fs.symlinkSync(binaryPath, aliasPath);

    console.log("migraine CLI installed successfully!");
    console.log('You can now use "migraine" or "mig" command.');
  } catch (error) {
    console.error("Error installing migraine:", error);
    process.exit(1);
  }
}

install();
