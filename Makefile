TARGET=connstat

$(shell rm -rf $(TARGET))
all: $(TARGET)
#--------------------------------------------------------------------------------------------------

$(TARGET):
	go build connstat.go
#--------------------------------------------------------------------------------------------------
clean:
	rm -rf $(TARGET)
