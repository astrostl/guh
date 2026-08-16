class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.3.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.3.0/guh-v0.3.0-darwin-arm64.tar.gz"
    sha256 "9f06a54efe30a90669da2dafbe41bd282013d4bc18b36ed8a19967a5100e7a47"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.3.0/guh-v0.3.0-darwin-amd64.tar.gz"
    sha256 "30d89fee6ae6f2f699fe765d1ffbc3ba79078fea23e34581ea558dc975c25f6a"
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
