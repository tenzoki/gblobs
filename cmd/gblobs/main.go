
package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "github.com/example/gblobs/gblobs"
)

// Usage: gblobs <command> ...
func main() {
    if len(os.Args) < 2 {
        fmt.Printf("gblobs <command> [options]\n")
        fmt.Println("Commands: putfile, putstring, get, exists, delete, purge, stats, inspect, search")
        os.Exit(1)
    }
    cmd := os.Args[1]
    switch cmd {
    case "putfile":
        putFileCmd(os.Args[2:])
    case "putstring":
        putStringCmd(os.Args[2:])
    case "get":
        getCmd(os.Args[2:])
    case "exists":
        existsCmd(os.Args[2:])
    case "delete":
        deleteCmd(os.Args[2:])
    case "purge":
        purgeCmd(os.Args[2:])
    case "stats":
        statsCmd(os.Args[2:])
    case "inspect":
        inspectCmd(os.Args[2:])
    case "search":
        searchCmd(os.Args[2:])
    default:
        fmt.Println("Unknown command.")
        os.Exit(1)
    }
}

// helpers for store setup
func openStoreOrDie(path, key string) *gblobs.LocalStore {
    st := &gblobs.LocalStore{}
    var err error
    if key != "" {
        err = st.OpenStore(path, key)
    } else {
        err = st.OpenStore(path)
    }
    if err != nil {
        fmt.Printf("Error opening store: %v\n", err)
        os.Exit(1)
    }
    return st
}

func openOrCreateStoreOrDie(path, key string) *gblobs.LocalStore {
    st := &gblobs.LocalStore{}
    var err error

    // Check if store directory exists
    if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
        // Directory doesn't exist, create store
        if key != "" {
            err = st.CreateStore(path, key)
        } else {
            err = st.CreateStore(path)
        }
        if err != nil {
            fmt.Printf("Error creating store: %v\n", err)
            os.Exit(1)
        }
    } else {
        // Directory exists, try to open store
        if key != "" {
            err = st.OpenStore(path, key)
        } else {
            err = st.OpenStore(path)
        }
        if err != nil {
            fmt.Printf("Error opening store: %v\n", err)
            os.Exit(1)
        }
    }
    return st
}

func createStoreOrDie(path, key string) *gblobs.LocalStore {
    st := &gblobs.LocalStore{}
    var err error
    if key != "" {
        err = st.CreateStore(path, key)
    } else {
        err = st.CreateStore(path)
    }
    if err != nil {
        fmt.Printf("Error creating store: %v\n", err)
        os.Exit(1)
    }
    return st
}

func putFileCmd(args []string) {
    fs := flag.NewFlagSet("putfile", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    owner := fs.String("owner", "", "Owner (optional)")
    fs.Parse(args)
    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs putfile <file> [flags]")
        os.Exit(1)
    }
    fname := fs.Arg(0)
    dat, err := os.ReadFile(fname)
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        os.Exit(1)
    }
    meta := gblobs.BlobType{
        Name: filepath.Base(fname),
        URI:  fname,
        Owner: *owner,
        IngestionTime: gblobs.NowUTC(),
    }
    st := openOrCreateStoreOrDie(*storePath, *key)
    id, err := st.PutBlob(dat, meta)
    if err != nil {
        fmt.Printf("Store error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(id)
}

func putStringCmd(args []string) {
    fs := flag.NewFlagSet("putstring", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    name := fs.String("name", "string", "Name of blob")
    owner := fs.String("owner", "", "Owner (optional)")
    fs.Parse(args)
    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs putstring <string> [flags]")
        os.Exit(1)
    }
    str := fs.Arg(0)
    meta := gblobs.BlobType{
        Name: *name,
        URI:  "string://local",
        Owner: *owner,
        IngestionTime: gblobs.NowUTC(),
    }
    st := openOrCreateStoreOrDie(*storePath, *key)
    id, err := st.PutBlob([]byte(str), meta)
    if err != nil {
        fmt.Printf("Store error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(id)
}

func getCmd(args []string) {
    fs := flag.NewFlagSet("get", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    outFile := fs.String("out", "", "Write output to file (default: stdout)")
    fs.Parse(args)
    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs get <blobID> [flags]")
        os.Exit(1)
    }
    blobID := fs.Arg(0)
    st := openStoreOrDie(*storePath, *key)
    data, meta, err := st.GetBlob(blobID)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    if *outFile == "" {
        if len(data) > 1024*1024 {
            fmt.Fprintf(os.Stderr, "Blob is large (%d bytes), writing metadata:\n", len(data))
            fmt.Printf("%+v\n", meta)
        }
        os.Stdout.Write(data)
    } else {
        if err := os.WriteFile(*outFile, data, 0o644); err != nil {
            fmt.Printf("Write error: %v\n", err)
            os.Exit(1)
        }
    }
}

func existsCmd(args []string) {
    fs := flag.NewFlagSet("exists", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    fs.Parse(args)
    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs exists <blobID> [flags]")
        os.Exit(1)
    }
    blobID := fs.Arg(0)
    st := openStoreOrDie(*storePath, *key)
    exists, err := st.ExistsBlob(blobID)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(exists)
}

func deleteCmd(args []string) {
    fs := flag.NewFlagSet("delete", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    fs.Parse(args)
    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs delete <blobID> [flags]")
        os.Exit(1)
    }
    blobID := fs.Arg(0)
    st := openStoreOrDie(*storePath, *key)
    err := st.DeleteBlob(blobID)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("deleted")
}

func purgeCmd(args []string) {
    fs := flag.NewFlagSet("purge", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    fs.Parse(args)
    st := openStoreOrDie(*storePath, *key)
    err := st.PurgeStore()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("purged store")
}

func statsCmd(args []string) {
    fs := flag.NewFlagSet("stats", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    fs.Parse(args)
    st := openStoreOrDie(*storePath, *key)
    stats, err := st.Stats()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Total blobs: %d\n", stats.TotalBlobCount)
    fmt.Printf("Max count per level: %v\n", stats.MaxCountPerLevel)
    fmt.Printf("Average count per level: %.2f\n", stats.AverageCountPerLevel)
}

func inspectCmd(args []string) {
    fs := flag.NewFlagSet("inspect", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    fs.Parse(args)
    st := openStoreOrDie(*storePath, *key)
    blobs, err := st.InspectStore()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    // Sort blobs by name first, then by IngestionTime (newest first)
    sort.Slice(blobs, func(i, j int) bool {
        if blobs[i].Name == blobs[j].Name {
            return blobs[i].IngestionTime.After(blobs[j].IngestionTime)
        }
        return blobs[i].Name < blobs[j].Name
    })

    fmt.Printf("Found %d blobs:\n", len(blobs))
    for _, blob := range blobs {
        fmt.Printf("%s [%s]\n", blob.Name, blob.BlobHash)
        fmt.Printf("  %s, %d bytes\n", blob.IngestionTime.Format("2006-01-02 15:04:05 UTC"), blob.Length)
        fmt.Printf("  URI: %s Owner: %s\n", blob.URI, blob.Owner)
        fmt.Println()
    }
}

func searchCmd(args []string) {
    fs := flag.NewFlagSet("search", flag.ExitOnError)
    storePath := fs.String("store", "./store", "Store directory")
    key := fs.String("key", "", "Encryption key (optional)")
    limit := fs.Int("limit", 10, "Maximum results to return")
    offset := fs.Int("offset", 0, "Starting offset for pagination")
    highlight := fs.Bool("highlight", true, "Show highlighted matches")
    fs.Parse(args)

    if fs.NArg() < 1 {
        fmt.Println("Usage: gblobs search <query> [flags]")
        fmt.Println("Flags:")
        fmt.Println("  --store <path>      Store directory (default: ./store)")
        fmt.Println("  --key <key>         Encryption key (optional)")
        fmt.Println("  --limit <n>         Maximum results (default: 10)")
        fmt.Println("  --offset <n>        Starting offset (default: 0)")
        fmt.Println("  --highlight         Show highlighted matches (default: true)")
        os.Exit(1)
    }

    query := fs.Arg(0)
    st := openStoreOrDie(*storePath, *key)

    // Prepare search request
    req := gblobs.SearchRequest{
        Query:     query,
        Limit:     *limit,
        Offset:    *offset,
        Highlight: *highlight,
    }

    results, err := st.SearchWithOptions(req)
    if err != nil {
        fmt.Printf("Search error: %v\n", err)
        os.Exit(1)
    }

    if len(results) == 0 {
        fmt.Printf("No results found for query: \"%s\"\n", query)
        return
    }

    fmt.Printf("Found %d results for \"%s\":\n\n", len(results), query)
    for i, result := range results {
        fmt.Printf("%d. %s (score: %.3f)\n",
            i+1, result.Metadata.Name, result.Score)
        fmt.Printf("   Blob ID: %s\n", result.BlobID)
        fmt.Printf("   %s\n", result.Metadata.URI)
        fmt.Printf("   %d bytes, %s",
            result.Metadata.Length,
            result.Metadata.IngestionTime.Format("2006-01-02 15:04:05"))

        if result.Metadata.Owner != "" {
            fmt.Printf(", owner: %s", result.Metadata.Owner)
        }
        fmt.Println()

        if *highlight && len(result.Highlights) > 0 {
            for field, fragments := range result.Highlights {
                if len(fragments) > 0 {
                    fmt.Printf("   %s: %s\n", field, strings.Join(fragments, " ... "))
                }
            }
        }
        fmt.Println()
    }
}
