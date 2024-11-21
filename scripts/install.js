const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const version = process.env.npm_package_version;
const platform = process.platform;
const arch = process.arch === "x64" ? "amd64" : process.arch;

// Skip installation if platform is not supported
if (platform !== "darwin" && platform !== "linux") {
  console.error(
    "Unsupported platform. Migraine CLI only supports macOS and Linux.",
  );
  process.exit(1);
}

// Skip installation if architecture is not supported
if (arch !== "amd64" && arch !== "arm64") {
  console.error(
    "Unsupported architecture. Migraine CLI only supports x64 and arm64.",
  );
  process.exit(1);
}

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
        if (response.statusCode === 302) {
          // Handle redirect
          https
            .get(response.headers.location, (redirectedResponse) => {
              redirectedResponse.pipe(file);
              file.on("finish", () => {
                file.close();
                resolve();
              });
            })
            .on("error", (err) => {
              fs.unlink(dest, () => reject(err));
            });
        } else {
          response.pipe(file);
          file.on("finish", () => {
            file.close();
            resolve();
          });
        }
      })
      .on("error", (err) => {
        fs.unlink(dest, () => reject(err));
      });
  });
};

async function install() {
  try {
    console.log("Installing Migraine CLI...");

    // Create bin directory
    const binPath = path.join(__dirname, "..", "bin");
    if (!fs.existsSync(binPath)) {
      fs.mkdirSync(binPath, { recursive: true });
    }

    const binaryPath = path.join(binPath, "migraine");
    const aliasPath = path.join(binPath, "mig");
    const downloadUrl = getDownloadUrl();

    console.log(`Downloading from: ${downloadUrl}`);
    await download(downloadUrl, binaryPath);

    // Make binary executable
    fs.chmodSync(binaryPath, "755");

    // Create symbolic link for alias
    if (fs.existsSync(aliasPath)) {
      fs.unlinkSync(aliasPath);
    }
    fs.symlinkSync(binaryPath, aliasPath);

    console.log("\n✨ Migraine CLI installed successfully!");
    console.log('You can now use either "migraine" or "mig" commands.\n');

    // Test the installation
    try {
      const version = execSync(`${binaryPath} --version`).toString().trim();
      console.log(`Installed version: ${version}`);
    } catch (error) {
      console.log('Note: Run "migraine --version" to verify the installation.');
    }
  } catch (error) {
    console.error("\n❌ Error installing Migraine CLI:", error.message);
    process.exit(1);
  }
}

// Run installation
install().catch((error) => {
  console.error("\n❌ Installation failed:", error.message);
  process.exit(1);
});
