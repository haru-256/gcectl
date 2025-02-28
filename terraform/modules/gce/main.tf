module "project_services" {
  source  = "terraform-google-modules/project-factory/google//modules/project_services"
  version = "~> 18.0"

  project_id = var.project_id
  activate_apis = [
    "compute.googleapis.com",
  ]
}

# VPC Network
resource "google_compute_network" "vpc_network" {
  name                    = "sandbox"
  auto_create_subnetworks = false
  project                 = module.project_services.project_id
}

# Subnetwork
resource "google_compute_subnetwork" "subnet" {
  name                     = "sandbox"
  ip_cidr_range            = "10.0.0.0/24"
  project                  = module.project_services.project_id
  region                   = var.region
  network                  = google_compute_network.vpc_network.name
  private_ip_google_access = true # Enable private Google access for Cloud NAT
}

# # Cloud Router
# resource "google_compute_router" "nat_router" {
#   name    = "nat-router"
#   project                 = module.project_services.project_id
#   region  = var.region
#   network = google_compute_network.vpc_network.name
# }

# # Cloud NAT
# resource "google_compute_router_nat" "cloud_nat" {
#   name                               = "cloud-nat"
#   router                             = google_compute_router.nat_router.name
#   project                 = module.project_services.project_id
#   region                             = var.region
#   nat_ip_allocate_option             = "AUTO_ONLY" # Automatically allocate external IPs for NAT
#   source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
# }

# # Firewall Rule for SSH (Optional)
# resource "google_compute_firewall" "iap_ssh_firewall" {
#   name      = "allow-ssh-from-iap"
#   project                 = module.project_services.project_id
#   network   = google_compute_network.vpc_network.name
#   direction = "INGRESS"

#   allow {
#     protocol = "tcp"
#     ports    = ["22"]
#   }

#   source_ranges = ["35.235.240.0/20"]
#   target_tags   = ["iap-ssh"]
# }

# Compute Instance (VM)
resource "google_compute_instance" "vm_instance" {
  name         = var.vm_name
  machine_type = "f1-micro"
  project      = module.project_services.project_id
  zone         = var.zone

  boot_disk {
    auto_delete = true
    initialize_params {
      image = "ubuntu-minimal-2410-oracular-arm64-v20250212"
      size  = 10
    }
  }

  network_interface {
    network    = google_compute_network.vpc_network.name
    subnetwork = google_compute_subnetwork.subnet.name
    # No external IP assigned to the VM
  }

  scheduling {
    provisioning_model          = "SPOT"
    preemptible                 = true
    automatic_restart           = false
    instance_termination_action = "STOP" # Optional: Specify action on preemption
  }

  metadata = {
    enable-oslogin = "TRUE"
  }

  tags = ["iap-ssh"] # Tag to apply the firewall rule for SSH access via IAP
}
