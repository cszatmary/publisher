{
  message: "Publish ${SHA} - ${DATE}"
  files: [
    "build/*"
    "README.md"
  ]
  preRun: "yarn build"
  targets: {
    production: {
      branch: "gh-pages"
      repo: "<org>/<name>"
      url: "custom.example.com"
    }
  }
}
