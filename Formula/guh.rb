class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.1.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.1.0/guh-v0.1.0-darwin-arm64.tar.gz"
    sha256 "4bec809463b87cf4d4802036745c3557342be03f65e14ca49fc710ae7798077d"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.1.0/guh-v0.1.0-darwin-amd64.tar.gz"
    sha256 "bf60046eea50d79922d5769917332ebdd61aeeb28e740a328754938e05a04041"
  else
    odie "guh is only supported on macOS via Homebrew. Build from source for Linux."
  end

  def install
    bin.install "guh-darwin-arm64" => "guh" if Hardware::CPU.arm?
    bin.install "guh-darwin-amd64" => "guh" if Hardware::CPU.intel?
  end

  test do
    system bin/"guh", "-version"
  end
end
