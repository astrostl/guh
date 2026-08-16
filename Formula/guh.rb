class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.6.1"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.6.1/guh-v0.6.1-darwin-arm64.tar.gz"
    sha256 "e1feba78e12b99a0cbe9bf6a76dd567f6aa8ec47a44cef6b7bcfeb2f24ee3f64"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.6.1/guh-v0.6.1-darwin-amd64.tar.gz"
    sha256 "cccb1fcfa4069e7dc668095e9a2cfda4ffff51509b787094faa802a4e766b6e0"
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
