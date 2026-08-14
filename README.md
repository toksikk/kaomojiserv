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

Seasonal presentations follow the visitor's local calendar. They can be previewed
outside their normal date ranges with `?season=easter`, `?season=april-fools`,
`?season=halloween`, `?season=christmas`, or `?season=new-year`. Use
`?season=none` to disable seasonal presentation during an active period.
