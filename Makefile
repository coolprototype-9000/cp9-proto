GOFLAGS = -race
FMTFLAGS = -s -w

TARGET = cp9psrv

.PHONY: all $(TARGET)

all: format $(TARGET)

$(TARGET):
	go build $(GOFLAGS) -o $(TARGET)

run: $(TARGET)
	./$(TARGET)

.PHONY: clean format
clean:
	rm -rf $(TARGET)

format:
	gofmt $(FMTFLAGS) .
