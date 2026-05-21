class Wtf < Formula
  desc "Where's The File? Blazing-fast interactive terminal file finder and CLI searcher"
  homepage "https://github.com/hariharen9/wtf"
  version "0.0.1"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/hariharen9/wtf/releases/download/v#{version}/wtf-darwin-arm64.tar.gz"
      # sha256 "DYNAMICALLY_POPULATED_BY_RELEASE_PIPELINE"
    else
      url "https://github.com/hariharen9/wtf/releases/download/v#{version}/wtf-darwin-amd64.tar.gz"
      # sha256 "DYNAMICALLY_POPULATED_BY_RELEASE_PIPELINE"
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/hariharen9/wtf/releases/download/v#{version}/wtf-linux-amd64.tar.gz"
      # sha256 "DYNAMICALLY_POPULATED_BY_RELEASE_PIPELINE"
    end
  end

  def install
    bin.install "wtf"
  end

  test do
    assert_match "wtf version #{version}", shell_output("#{bin}/wtf -v")
  end
end
