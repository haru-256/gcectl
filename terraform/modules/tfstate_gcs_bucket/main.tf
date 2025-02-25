# allow to use the google cloud service api
resource "google_project_service" "services" {
  project            = var.gcp_project_id
  service            = "storage.googleapis.com"
  disable_on_destroy = true
}

resource "google_storage_bucket" "tfstate_bucket" {
  name                        = "${var.gcp_project_id}-tfstate"
  location                    = "US"
  force_destroy               = true
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  depends_on                  = [google_project_service.services]
  autoclass {
    enabled                = true
    terminal_storage_class = "NEARLINE"
  }
  lifecycle_rule {
    condition {
      num_newer_versions = 3
    }
    action {
      type = "Delete"
    }
  }
  versioning {
    enabled = true
  }
}
