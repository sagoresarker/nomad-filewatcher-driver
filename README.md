# Nomad File Watcher Driver

A Nomad task driver that monitors file system changes and executes commands based on events.

## Features

- Monitor multiple directories and files
- Execute commands on file system events
- Support for recursive directory watching
- Pattern-based file/directory ignoring
- Environment variable passing
- State persistence
- Metrics exposure

## Installation

1. Build the plugin:
```bash
./scripts/build.sh
```

2. Install the plugin:
```bash
sudo ./scripts/install.sh
```

## Configuration

### Client Configuration

Add to your Nomad client configuration:

```hcl
plugin "filewatcher-driver" {
  config {
    enabled = true
    state_dir = "/var/lib/nomad/filewatcher"
    log_level = "INFO"
  }
}
```

### Job Configuration

Example job specification:

```hcl
job "file-monitor" {
  // See example/jobs/watcher-job.nomad for full example
}
```

## Usage

1. Start Nomad with the plugin enabled
2. Submit a job:
```bash
nomad job run example/jobs/watcher-job.nomad
```

3. Monitor the job:
```bash
nomad status file-monitor
nomad alloc logs <alloc_id> log-watcher
```

## License

MIT