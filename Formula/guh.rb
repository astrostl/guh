class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.9.3"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.9.3/guh-v0.9.3-darwin-arm64.tar.gz"
    sha256 "e2efb6b607b51adef40e0b1c6a96d81212f36437916f34ead465c66add9bae59"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.9.3/guh-v0.9.3-darwin-amd64.tar.gz"
    sha256 "05325d7183dd8dc3cbac0e241a82b7cc732e84e69c2c6646d7a9717d4b13eb5b"
  else
    odie "guh is only supported on macOS via Homebrew. Build from source for Linux."
  end

  def install
    bin.install "guh-darwin-arm64" => "guh" if Hardware::CPU.arm?
    bin.install "guh-darwin-amd64" => "guh" if Hardware::CPU.intel?
  end

  test do
    system bin/"guh", "--version"
  end
end
