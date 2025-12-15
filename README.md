commands
1. ruborag parse [file1] [file2] -strips html tags
2. ruborag embed [file1] - embeds the content of file into vector embeddings
3. ruborag search "query" - use cosine similarity to search and find top k relevant docs 
4. ruborag ask "query" - use RAG and LLM to answer query based on rust book



## Benchmarks

### ruborag parse
1. Unbuffered IO
```bash
go run main.go parse -w -u --out-dir=./corpus/parsed corpus/raw/*
```

2. Buffered IO
```bash
go run main.go parse -w --out-dir=./corpus/parsed corpus/raw/*
```

| IO strategy   | User time (s) | System time (s) | CPU usage | Total time (s) |
|--------------|---------------|-----------------|-----------|----------------|
| Unbuffered IO | 0.16          | 0.16            | 24%       | 1.282          |
| Buffered IO   | 0.09          | 0.12            | 56%       | 0.358          |

Buffered IO is ~3.5x faster than unbuffered IO.


