package main


import (
    "fmt"
    "net"
    "net/textproto"
    "log"
    "bufio"
    "strings"
    "time"
    "os"
    "encoding/json"
)

type event struct {
    code string
    message string
}

type config struct {
    Pass string
    Nick string
    Channel string
}

func main() {
    file, err := os.Open("config.json")

    if err != nil {
        log.Fatal("Error opening config file", err)
    }

    var config config
    decoder := json.NewDecoder(file) 
    err = decoder.Decode(&config)

    if err != nil {
        log.Fatal("Error loading config file", err)
    }

    file.Close()

    conn := connect(config)

    eventChan := make(chan event)
    go readLines(eventChan, conn)

    isOnCooldown := false
    
    for {
        event := <-eventChan
        go handleEvent(event, conn, config.Channel, &isOnCooldown)
    }
}

func handleEvent(event event, conn net.Conn, channel string, isOnCooldown *bool) {
    if event.code == "PING" {
        fmt.Fprintf(conn, "PONG :%s\r\n", event.message)
    } else if event.code == "PRIVMSG" {
        if !*isOnCooldown && strings.Contains(event.message, "!subhype") {
            fmt.Fprintf(conn, "PRIVMSG %s :coruscAww Bappa coruscAww Bappa coruscAww Bappa coruscAww\r\n", channel)
            go cooldown(isOnCooldown)
        }
    }
}

func cooldown(isOnCooldown *bool) {
    *isOnCooldown = true
    time.Sleep(time.Second * 30)
    *isOnCooldown = false
}

func parseEvent(line string) event {
    var event event

    if strings.HasPrefix(line, "@") {
        // TODO: do something with this instead of discarding
        line = line[strings.Index(line, " ") + 1:]
    }
    
    if strings.HasPrefix(line, ":") { 
        parts := strings.SplitN(strings.TrimPrefix(line, ":"), " :", 2)
        args := strings.Split(parts[0], " ")

        if len(parts) == 2 && len(args) > 1 {
            event.code = args[1]
            event.message = parts[1]
        }
    } else {
        parts := strings.SplitN(line, " :", 2)

        if len(parts) == 2 {
            event.code = parts[0]
            event.message = parts[1]
        }
    }
    
    return event
}

func readLines(eventChan chan event, conn net.Conn) {
    reader := bufio.NewReader(conn)
    tp := textproto.NewReader(reader) 

    for {
        line, err := tp.ReadLine()

        if err != nil {
            log.Fatal("Something went wrong", err)
            break
        }

        fmt.Printf("%s\n", line)
        eventChan<- parseEvent(line)
    }
}

func connect(config config) net.Conn {
    conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")

    if err != nil {
        log.Fatal("Unable to connect", err)
    }

    log.Printf("Connected")

    fmt.Fprintf(conn, "PASS %s\r\n", config.Pass)
    fmt.Fprintf(conn, "NICK %s\r\n", config.Nick)
    fmt.Fprintf(conn, "JOIN %s\r\n", config.Channel)
    fmt.Fprintf(conn, "CAP REQ :twitch.tv/membership\r\n")
    fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags\r\n")
    fmt.Fprintf(conn, "CAP REQ :twitch.tv/commands\r\n")

    return conn
}
