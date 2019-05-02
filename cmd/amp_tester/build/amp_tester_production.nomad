job "amp_tester" {
	region = "s2"
	datacenters = ["s2"]
	
	update {
		# canary = 1
		max_parallel = 1
		min_healthy_time = "5s"
		healthy_deadline = "30s"
		auto_revert = true
	}
	
	group "amp_tester" {

		task "amp_tester" {
			driver = "docker"
			
			service {
				name = "amp-tester"				
				port = "debug"
				check {
					type		 = "http"
					path		 = "/health_check"
					interval = "10s"
					timeout	 = "1s"
				}
			}
      service {
        name = "amp-tester-ws"
        port = "ws"
        check {
					type		 = "tcp"
					interval = "10s"
					timeout	 = "1s"
				}
      }
      service {
        name = "amp-tester-app"
        port = "app"
        check {
					type		 = "http"
					path		 = "/index.html"
					interval = "10s"
					timeout	 = "1s"
				}
      }
      service {
        name = "amp-tester-log"
        port = "log"
        check {
					type		 = "http"
					path		 = "/health_check"
					interval = "10s"
					timeout	 = "1s"
				}
      }

			config {
				image = "registry.dev.minus5.hr/amp_tester:v0.0.22"
				dns_servers = ["${attr.unique.network.ip-address}", "8.8.8.8"]
				hostname = "${node.unique.id}"

				logging {
					type = "syslog"
					config {
						syslog-address = "udp://${attr.unique.network.ip-address}:514"
						tag = "amp_tester"
						syslog-format = "rfc3164"
					}
				}
			}

			env {
				SVCKIT_DCY_CONSUL = "${attr.unique.network.ip-address}:8500"
				SVCKIT_NSQD = "${attr.unique.network.ip-address}:4150"
        SVCKIT_LOG_SYSLOG = "${attr.unique.network.ip-address}:514"
				STATSD_LOGGER_ADDRESS = "${attr.unique.network.ip-address}:18125"
				SVCKIT_LOG_DISABLE_DEBUG = 1
			}

			resources {
				cpu = 100
				memory = 384
				network {
					mbits = 1
					port "debug" {}
          port "ws" {}
          port "app" {}
          port "log" {}
				}
			}
		}
	}
}
