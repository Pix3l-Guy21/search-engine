# Local Search Engine

This is a local pdf search engine for specific words in the pdfs

## Project Structure

```text
search-engine/
├── cmd/
|   └── engine/
└── internal/
    ├── crawler/
    ├── indexer/
    ├── parser/
    ├── pipeline/
    └── store/
```

---

## How to use it

---

1. Clone the repo  

```bash
git clone https://github.com/Pix3l-Guy21/search-engine.git
```

2. In the root directory create *.env* file with the following format
 
 ```.env
 ROOT=your target directory
 ```

Repalace "your target directory" with the directory you want to implement the search engine at.

3. Run the following command in the terminal

 ```bash
 go run cmd/engine/main.go
 ```
