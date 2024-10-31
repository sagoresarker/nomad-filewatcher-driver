job "file-monitor" {
  datacenters = ["dc1"]
  type = "system"

  group "watchers" {
    volume "state" {
      type      = "host"
      source    = "filewatcher-state"
      read_only = false
    }

    task "log-watcher" {
      driver = "filewatcher-driver"

      volume_mount {
        volume      = "state"
        destination = "/var/lib/nomad/filewatcher"
        read_only   = false
      }

      config {
        paths = [
          "/var/log/syslog",
          "/var/log/auth.log",
          "/app/logs"
        ]

        events = ["create", "modify", "delete"]

        exec_command = "/usr/local/bin/alert-handler.sh"
        exec_args = [
          "-severity", "high",
          "-notify", "slack,email"
        ]

        recursive_watch = true

        ignore_patterns = [
          "*.tmp",
          "*.swp",
          ".git/*"
        ]

        environment = {
          ALERT_API_KEY = "secret-key",
          SLACK_WEBHOOK = "https://hooks.slack.com/services/xxx",
          NOTIFY_EMAIL = "admin@example.com"
        }
      }

      resources {
        cpu    = 200
        memory = 256
      }

      service {
        name = "file-monitor"
        port = "metrics"

        check {
          type     = "http"
          path     = "/metrics"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }

    task "metrics" {
      driver = "docker"

      config {
        image = "prom/node-exporter:latest"

        args = [
          "--collector.textfile.directory=/metrics"
        ]
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}