class Cuckoo < Formula
    desc "CLI Tool for GitLab CI and Kubernetes Deployments."
    url "https://circle-artifacts.com/gh/borchero/cuckoo/$CIRCLE_BUILD_NUM/artifacts/0/cuckoo.tar.gz"
    sha256 "$ARTIFACT_SHA256"

    depends_on "go@1.14" => :build

    def install
        bin.install "cuckoo"
    end

    test do
        cuckoo help
    end
end
