class Cuckoo < Formula
    desc "CLI Tool for GitLab CI and Kubernetes Deployments."
    url "<TODO>"
    sha256 "<TODO>"

    depends_on "go@1.14" => :build

    def install
        system "cd source && go build -v"
        bin.install "source/cuckoo" => "cuckoo"
    end

    test do
        cuckoo help
    end
end
