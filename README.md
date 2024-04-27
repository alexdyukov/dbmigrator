# dbmigrator
Go database migration library compatible with sql.DB
====
[![GoDoc](https://godoc.org/github.com/alexdyukov/dbmigrator?status.svg)](https://godoc.org/github.com/alexdyukov/dbmigrator)
[![Tests](https://github.com/alexdyukov/dbmigrator/actions/workflows/tests.yml/badge.svg?branch=master)](https://github.com/alexdyukov/dbmigrator/actions/workflows/tests.yml?query=branch%3Amaster)

This package provides method for database migration any stdlib's sql.DB compatibility database

## Restrictions

* Library does not care about version naming. Migrations runs in a lexical sort manner of filenames

  Semver/integer/timestamp typed versions requires different parse logics. Best run order was invented in unix's init.d : do lexical sort of files. Nothing good invented after that thing

- Ensure you trust fs.FS content. Library executes it as is, without any validation

  Library does not care of migration folder content. It does not know any DB engine specific functions/syntax

- Do not use transaction in migrations

  Library executes migration inside transactions to commit success upgrade in version table. There is no way to embed transaction in transaction

- There is no downgrade feature and never release it

  Because once you upgraded it, no one can guarantee that other application instances wont downgrade it. That means you should do version check every transaction. I have never seen such braindead devs who wrote code like this

- Try to migrate safety

  Do not break DB schemes by 1 migration. Use common way with at least 2 releases:
  1. release app version N+1 with new scheme without any breaking changes
  2. wait until rolling update completes
  3. release app version N+2 with removes unused parts of scheme

## License

MIT licensed. See the included LICENSE file for details.