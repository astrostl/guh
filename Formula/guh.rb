class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.2.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.2.0/guh-v0.2.0-darwin-arm64.tar.gz"
    sha256 "4a468efcabd5a672051d633c96aba0b482b7707c6d026e310a593f45d495b604"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.2.0/guh-v0.2.0-darwin-amd64.tar.gz"
    sha256 "b4fcf86a25edf2ae3bcc07a54d10937789a4d22dcd9593f19d80a26c586b936c"
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
