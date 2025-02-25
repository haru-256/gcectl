provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_default_region
}

provider "google-beta" {
  project = var.gcp_project_id
  region  = var.gcp_default_region
}
