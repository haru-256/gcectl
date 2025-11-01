locals {
  vm_names = ["sandbox-1", "sandbox-2"]
  # このTerraform構成で必要な全APIをリスト化
  required_services = [
    "compute.googleapis.com", # VMモジュール用
    "storage.googleapis.com", # GCSモジュール用
  ]
}

# 必要なAPIをすべて有効化し待機
module "required_project_services" {
  source = "../../modules/google_project_services"

  project_id        = var.gcp_project_id
  required_services = local.required_services
  wait_seconds      = 30
}

# google cloud project
data "google_project" "project" {
  project_id = var.gcp_project_id
}

# create the bucket for terraform state
module "tfstate_bucket" {
  source         = "../../modules/tfstate_gcs_bucket"
  gcp_project_id = data.google_project.project.project_id
}

module "sandbox_vms" {
  source             = "../../modules/gce"
  project_id         = var.gcp_project_id
  region             = var.gcp_default_region
  zone               = var.gcp_default_zone
  machine_type       = "f1-micro"
  vm_names           = local.vm_names
  with_stop_schedule = true
}
