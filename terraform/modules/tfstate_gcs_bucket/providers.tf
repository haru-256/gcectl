terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~>7.35.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~>7.30.0"
    }
  }
}
