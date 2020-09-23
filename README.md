# Mrunner: tool to run container workloads with different kata configs

```bash
git clone https://github.com/jcvenegas/mrunner.git
go build
# Use sudo if your user has not permissiones to use docker client.
sudo -E PATH=$PATH ./mrunner
```
What it can do?
- Run a test on top of docker
- Run the same test for multiple kata configs
  - kerne path
  - virtiofsd cache type
  - virtiofsd dax size
  - kata-runtime (based in paths from kata-deploy)
  - Extra virtiofs args

It will generate directory named: `results`

Workloads:
- For now just modify the workload in `main.go`

Example:

```
results/
├── kata-clh-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs
│   ├── fio.json
│   ├── kata-configuration.toml
│   ├── kata-env.json
│   └── result.json
└── kata-qemu-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs
    ├── fio.json
    ├── kata-configuration.toml
    ├── kata-env.json
    └── result.json
```

2 directories, 8 files


TODO:
- Create workloads definitions in yaml
