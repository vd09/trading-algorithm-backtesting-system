FROM grafana/grafana:latest

# Copy the provisioning files
COPY provisioning /etc/grafana/provisioning

# Copy the dashboards
COPY provisioning/dashboards /var/lib/grafana/dashboards
