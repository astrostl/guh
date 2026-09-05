class Guh < Formula
  desc "TUI for GitHub repos via the local gh session"
  homepage "https://github.com/astrostl/guh"
  version "v0.7.0"
  license "MIT"

  depends_on "gh"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/astrostl/guh/releases/download/v0.7.0/guh-v0.7.0-darwin-arm64.tar.gz"
    sha256 "c828065701a9b227cda63d5516b0bbd97ee4f3eafa9b0cc63823414c582529ac"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/astrostl/guh/releases/download/v0.7.0/guh-v0.7.0-darwin-amd64.tar.gz"
    sha256 "72c2b375d59e2d677e1762b407ff692895e7cb5623ca5a7a36eb1f47091e3253"
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
