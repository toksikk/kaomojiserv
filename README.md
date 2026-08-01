# kaomojiserv ( ˘▽˘)っ♨

This small tools purpose is serving Kaomojis to visitors.
They can use them as entertainment, decision support or basically anything they desire - free of charge!
More Kaomojis can be added to `kaomojis.txt`.

## Development

Install `pre-commit` and `golangci-lint`, then enable the repository hooks:

```sh
pre-commit install --hook-type pre-commit --hook-type commit-msg
```

The hooks format Go files, run linting and tests, verify modules, validate YAML and
repository hygiene, and enforce conventional commit messages such as
`fix(lint): check errors`.
