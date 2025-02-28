output "vm_instance_id" {
  value       = google_compute_instance.vm_instance.id
  description = "The ID of Google Cloud Compute Engine instance"
}
