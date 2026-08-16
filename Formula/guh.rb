class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.6.2"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.6.2/guh-v0.6.2-darwin-arm64.tar.gz"
    sha256 "a54f262b75c5f451790682e5d2dadd02e0fa13deea7c7a9b17c761f71bd87d3b"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.6.2/guh-v0.6.2-darwin-amd64.tar.gz"
    sha256 "1ad77e6016de6a0ae42923c2b8566c3184cd8e7f5eca9b8799ac6dbf3bec1240"
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
