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
user   0.16s
system 0.16s
cpu    24%
total  1.282s

2. Buffered IO
```bash
go run main.go parse -w --out-dir=./corpus/parsed corpus/raw/*
user   0.09s
system 0.12s
cpu    56%
total  0.358s
```



