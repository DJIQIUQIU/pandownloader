package elk

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"

    "git-sec.com/pandownloader/logger"
    "git-sec.com/pandownloader/module/utils"
)


type BaseRecord struct {
    EventName string `json:"event_name"`
    SystemCode string `json:"system_code"`
    EventType string `json:"event_type"`
}

type FileUploadedRecord struct {
    BaseRecord
    FileId int `json:"fileid"`
    User string `json:"user"`
    IP string `json:"ip"`
    FilePath string `json:filepath`
}


type DownloadFileRecord struct {
    BaseRecord
    FileId int `json:"fileid"`
    User string `json:"user"`
    IP string `json:"ip"`
    FilePath string `json:"filepath"`
    FileSize int `json:"filesize"`
}


func NewFileUploadedRecord(fileId int, user string, ip string, filePath string) FileUploadedRecord{
    return FileUploadedRecord{
        BaseRecord: BaseRecord{
            EventName: "pan_record",
            EventType: "file_upload",
            SystemCode: "",
        },
        FileId: fileId,
        User: user,
        IP: ip,
        FilePath: filePath,
    }
}


func NewDownloadFileRecord(fileId int, user string, ip string, filePath string, size int) DownloadFileRecord {
    return DownloadFileRecord{
        BaseRecord: BaseRecord{
            EventName: "pan_download_record",
            EventType: "",
            SystemCode: "",
        },
        FileId: fileId,
        User: user,
        IP: ip,
        FilePath: filePath,
        FileSize: size,
    }
}


func (record FileUploadedRecord) Send() error {
    b, err := json.Marshal(&record)
    if err != nil {
        logger.GetLogger().Println("Send to ELK Error: ", err)
        return err
    }
    return send2elk(b)
}


func (record DownloadFileRecord) Send() error {
    b, err := json.Marshal(&record)
    if err != nil {
        logger.GetLogger().Println("Send to ELK Error: ", err)
        return err
    }
    return send2elk(b)
}


func send2elk(data []byte) error {
    write2elk, _ := utils.Cfg.GetValue("log", "write2elk")
    if write2elk != "1" {
        logger.GetLogger().Println("Do not send to ELK.")
        return nil
    }
    url, _ := utils.Cfg.GetValue("log", "elk")
    c := &http.Client{
        Timeout: 15 * time.Second,
    }

    body := bytes.NewBuffer(data)
    logger.GetLogger().Printf("Send to ELK: %v", body.String())
    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        logger.GetLogger().Println("Send to ELK Error: ", err)
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.Do(req)
    if err != nil {
        logger.GetLogger().Println("Send to ELK Error: ", err)
        return err
    }
    logger.GetLogger().Printf("Send to ELK %v: %d", data, resp.StatusCode)
    return nil
}
