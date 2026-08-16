class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.5.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.5.0/guh-v0.5.0-darwin-arm64.tar.gz"
    sha256 "fdf3e13d587b72fc1d7ed45d8f678d29a6d1f510f0c93d9eed749da6fc16aec2"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.5.0/guh-v0.5.0-darwin-amd64.tar.gz"
    sha256 "c4ef1baa7e3f1c6c7da8851381c2b35c7e2fdb9c4a0c450e7e6d5178363ef019"
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
