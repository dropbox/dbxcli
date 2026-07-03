class Dbxcli < Formula
  desc "Command-line tool for Dropbox users and team admins"
  homepage "https://github.com/dropbox/dbxcli"
  url "https://github.com/dropbox/dbxcli/archive/refs/tags/v3.6.0.tar.gz"
  sha256 "49d80ff75f879420ae0e20bd77172a1435edd7e15bf3068cfbe5696d89a8c43b"
  license "Apache-2.0"
  head "https://github.com/dropbox/dbxcli.git", branch: "master"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
    generate_completions_from_executable bin/"dbxcli", "completion", shells: [:bash, :zsh, :fish]
    doc.install Dir["docs/commands/*.md"]
  end

  test do
    ENV["DBXCLI_AUTH_FILE"] = testpath/"missing-auth.json"
    assert_path_exists doc/"dbxcli_completion.md"
    assert_path_exists doc/"dbxcli_completion_bash.md"
    output = shell_output("#{bin}/dbxcli ls 2>&1", 2)
    assert_match "no saved Dropbox credentials", output
  end
end
