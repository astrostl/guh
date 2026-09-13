class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.9.2"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.9.2/guh-v0.9.2-darwin-arm64.tar.gz"
    sha256 "d784afa13dbd1173e161d654f7c14d0f56fd2f256238fb975dd4f429ec825a7b"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.9.2/guh-v0.9.2-darwin-amd64.tar.gz"
    sha256 "17c7627a3fb11832933f34756414f0b7eb761755a7ad56c70cc7584ee0891a8e"
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
