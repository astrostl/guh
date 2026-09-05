class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.9.1"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.9.1/guh-v0.9.1-darwin-arm64.tar.gz"
    sha256 "3e82aeb8c80d593691a82fb16073120487855cbd0cd76ba6d048c04c0aa14786"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.9.1/guh-v0.9.1-darwin-amd64.tar.gz"
    sha256 "8fa1fd57ae7adf2bb6518a3980f04fb23665dfd1d1efe574a692b3b37afad146"
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
