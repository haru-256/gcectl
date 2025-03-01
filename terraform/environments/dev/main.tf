# google cloud project
data "google_project" "project" {
  project_id = var.gcp_project_id
}

# create the bucket for terraform state
module "tfstate_bucket" {
  source         = "../../modules/tfstate_gcs_bucket"
  gcp_project_id = data.google_project.project.project_id
}

module "sandbox_vm" {
  source             = "../../modules/gce"
  project_id         = var.gcp_project_id
  region             = var.gcp_default_region
  zone               = var.gcp_default_zone
  machine_type       = "f1-micro"
  vm_name            = "sandbox"
  with_stop_schedule = true
}
