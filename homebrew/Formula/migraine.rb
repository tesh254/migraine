class Migraine < Formula
    desc "migraine is a command-line interface (CLI) tool designed to help you manage migrations in your backend project using a PostgreSQL database."
    homepage "https://github.com/tesh254/migraine"
    url "https://github.com/tesh254/migraine/releases/download/v0.0.2-alpha.5/migraine_0.0.2-alpha.5_darwin_amd64.tar.gz"
    sha256 "a6544b14388e58845358a51d80705fb5f6c8c00185288086376be1a4d4d698c3"
    version "v0.0.2-alpha.5"
    version_scheme 1
  
    on_macos do
      if Hardware::CPU.arm?
        url "https://github.com/tesh254/migraine/releases/download/v0.0.2-alpha.5/migraine_0.0.2-alpha.5_darwin_arm64.tar.gz"
        sha256 "63e74cad123c160361b0b2b87eb3254b6d47dd1f8534fd9c3c23f17a0e45e149"
      end
    end
  
    def install
      bin.install "migraine"
      bin.install_symlink "migraine" => "mg"
    end
  
    test do
      system bin/"migraine", "--version"
      system bin/"mg", "--version"
    end
end