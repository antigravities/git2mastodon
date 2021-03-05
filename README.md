# git2mastodon
Post new Git commits to Mastodon-compatible Fediverse instances

## Install and run
rss2discord *requires* Go 1.16+.

```
go get get.cutie.cafe/git2mastodon
```

## Usage
Run this as part of a cron job (or equivalent).

```
Usage of git2mastodon:
  -force
        Post a commit again even if it's already been posted.
  -instance string
        The Mastodon (or Mastodon-compatible) instance to interface with (only required on first run) (default "https://mastodon.social")
  -refspec string
        The refspec to compare commits with. (default "refs/heads/master")
  -repo string
        The repository to fetch.
  -run-every uint
        If > 0, git2mastodon will run in the foreground and run a check every X seconds.
  -storage string
        File to store settings/data in. (default "masto.cfg")
  -tmpl string
        A Go template used to encode a status to post. 
```

Example:
```
./git2mastodon -instance https://mstdn.social -repo https://github.com/tootsuite/mastodon
```

## License
```
Copyright (C) 2021 Alexandra Frock

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```