# example/config/client.hcl

client {
  enabled = true

  plugin "filewatcher-driver" {
    config {
      enabled = true

      # Plugin-specific settings
      state_dir = "/var/lib/nomad/filewatcher"
      log_level = "INFO"

      # Default settings for all watchers
      default_recursive = true
      max_watch_paths = 100
      event_buffer_size = 1000
    }
  }

  # Host volume for persistent storage
  host_volume "filewatcher-state" {
    path = "/var/lib/nomad/filewatcher"
    read_only = false
  }
}