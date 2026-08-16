class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.6.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.6.0/guh-v0.6.0-darwin-arm64.tar.gz"
    sha256 "e7ee4b16610653bbc4d55470c3034cebb4e4691bbee1123082d84cc06b9b3868"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.6.0/guh-v0.6.0-darwin-amd64.tar.gz"
    sha256 "4b79bbf03b36df1d75c3179eb77477c933550352ed06c1395bb45a1efbb954c9"
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
