class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.9.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.9.0/guh-v0.9.0-darwin-arm64.tar.gz"
    sha256 "3b9c89a146011a3db5e30e5d9e52896a1b29743c80863aa5e14de24399066f8d"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.9.0/guh-v0.9.0-darwin-amd64.tar.gz"
    sha256 "c0967a1072469656836d0557348d41e4f25b274bd05734a3a47caa20b2c0a9fe"
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
