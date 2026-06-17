class Dbxcli < Formula
  desc "Command-line tool for Dropbox users and team admins"
  homepage "https://github.com/dropbox/dbxcli"
  url "https://github.com/dropbox/dbxcli/archive/refs/tags/v3.3.3.tar.gz"
  sha256 "be0187b703ef726b21ace33212a9b9f743502e74a8149d2f356fda650408c1a7"
  license "Apache-2.0"
  head "https://github.com/dropbox/dbxcli.git", branch: "master"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
    generate_completions_from_executable bin/"dbxcli", "completion", shells: [:bash, :zsh, :fish]
  end

  test do
    assert_match "dbxcli version: #{version}", shell_output("#{bin}/dbxcli version")
  end
end
