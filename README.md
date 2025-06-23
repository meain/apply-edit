# apply-edit

A command-line tool that performs precise search-and-replace operations on text files using a special diff format.
See [aider diff format](https://aider.chat/docs/more/edit-formats.html#diff).

> Built to mostly use with [esa](https://github.com/meain/esa)

## Installation

```bash
go install github.com/meain/apply-edit@latest
```

Or clone the repository and build it manually:

```bash
git clone https://github.com/meain/apply-edit.git
cd apply-edit
go build
```

## Usage

```bash
apply-edit [--explain] <filename>
```

### Arguments

- `<filename>`: The target file to modify

### Options

- `--explain`: Display detailed usage information and examples

## Description

`apply-edit` reads a special diff format from stdin and applies the changes to the specified file. The diff consists of a search block and a replace block, allowing you to precisely target and modify text content.

The tool is designed to be:
- **Precise**: Only replaces exact matches, including whitespace
- **Safe**: Refuses to make changes if multiple matches are found, avoiding ambiguity
- **Simple**: Uses an intuitive diff format inspired by merge conflict markers

## Special Diff Format

The diff format uses markers similar to Git merge conflict markers:

```
<<<<<<< SEARCH
[text to find]
=======
[text to replace with]
>>>>>>> REPLACE
```

- The text between `<<<<<<< SEARCH` and `=======` is what will be searched for
- The text between `=======` and `>>>>>>> REPLACE` is what will replace the search text

## Examples

### Adding an Import Statement

File content (app.py):
```python
from flask import Flask
app = Flask(__name__)
```

Command:
```bash
cat <<EOF | apply-edit app.py
<<<<<<< SEARCH
from flask import Flask
=======
import math
from flask import Flask
>>>>>>> REPLACE
EOF
```

Result:
```python
import math
from flask import Flask
app = Flask(__name__)
```

### Removing Text

Command:
```bash
cat <<EOF | apply-edit app.py
<<<<<<< SEARCH
app = Flask(__name__)
=======
>>>>>>> REPLACE
EOF
```

This will delete the line `app = Flask(__name__)` from the file.

## Important Notes

- The search text must match exactly (including whitespace)
- If multiple matches exist, the operation will fail to avoid ambiguous edits
- Empty replace blocks will delete the search text
- The original file is overwritten with the changes
- Line endings are normalized during the search process