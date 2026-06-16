class Dbxcli < Formula
  desc "Command-line tool for Dropbox users and team admins"
  homepage "https://github.com/dropbox/dbxcli"
  url "https://github.com/dropbox/dbxcli/archive/refs/tags/v3.3.1.tar.gz"
  sha256 "729a50ba14301aff7610089a056d47344157628f182a7c7e31bde4cce935cfe2"
  license "Apache-2.0"
  head "https://github.com/dropbox/dbxcli.git", branch: "master"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
    generate_completions_from_executable bin/"dbxcli", "completion", shells: [:bash, :zsh, :fish]
  end

  test do
    assert_match "Usage:", shell_output("#{bin}/dbxcli --help")
  end
end
