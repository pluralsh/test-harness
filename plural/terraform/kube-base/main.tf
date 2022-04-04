resource "kubernetes_namespace" "test-harness" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/managed-by" = "plural"
      "app.plural.sh/name" = "test-harness"

    }
  }
}

