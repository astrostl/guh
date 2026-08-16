class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.4.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.4.0/guh-v0.4.0-darwin-arm64.tar.gz"
    sha256 "26a8b507fb4a2f00cf39d2138c509e7810c6d6460fbae2f3edecca666cceb97a"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.4.0/guh-v0.4.0-darwin-amd64.tar.gz"
    sha256 "22c0c8cf5a15d3352bc0af209a06c927853b290fee73b7d07993cb3c7000aebb"
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
