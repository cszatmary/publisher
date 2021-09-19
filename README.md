# publisher

publisher is a small CLI for publishing static sites to GitHub Pages.

## Usage

To publish a given target:

```sh
publisher -T <target>
```

```
Usage of publisher:
  -p, --path string     The path to the publisher.sc config file. (default "publisher.sc")
      --skip-prerun     Skip preRun step.
  -t, --tag string      The git tag to create. Omit if you do not want to create a tag.
  -T, --target string   The target to deploy.
  -v, --verbose         Enables verbose logging.
```

You can run `publisher -h` to view the help message.

## Configuration

To configure publisher create a `publisher.sc` file. An example is available in [`publisher.example.sc`](publisher.example.sc).

### `message: string`

The commit message to use. The following variables can be used:

- `${SHA}` - The SHA of the latest git commit.
- `${DATE}` - The current date formatted as `MM-DD-YYYY`.
- `${TAG}` - The Git tag being created that was supplied by the `--tag` flag.

### `files: string[]`

A list of files/directories to be published.

### `preRun: string`

A shell command to execute before publishing the site.
Example: `yarn build` to create a production build.

### `targets: map`

A map of targets that can be published to.
Publisher can be configured to publish to multiple GitHub repos in order to deploy different versions of a site.

Example:

```sc
targets: {
  production: {
    branch: "gh-pages"
    repo: "Example/my-site"
    url: "custom.example.com"
  }
}
```

A target has the following fields:

#### `branch: string`

The branch that is used for GitHub Pages. This is usually `gh-pages` or `master`.

#### `repo: string`

The GitHub repo to publish to. Must be of the form `<org>/<name>`.

#### `url: string`

A custom domain that the GitHub Pages site should use.
If provided, publisher will generate a `CNAME` file with this value.

## License

publisher is available under the [MIT License](LICENSE).

## Contributing

Open an issue or submit a pull request.
