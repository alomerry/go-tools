package mysql

import (
	"bufio"
	"fmt"
	"github.com/alomerry/go-tools/static/constant"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	mysqldump = "mysqldump"
)

type DumpTool struct{}

func (d *DumpTool) Dump(prefix string, params map[string]any, db constant.Database) (string, error) {
	var (
		date        = time.Now().Format(time.DateOnly)
		cmd         = exec.Command(mysqldump, append(d.genDumpCmdParam(params), db.Name)...)
		dumpSqlPath = fmt.Sprintf("%s/%s-%s.sql", prefix, db.Name, date)
	)

	dumpSql, err := os.OpenFile(dumpSqlPath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return constant.EmptyStr, err
	}
	defer dumpSql.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return constant.EmptyStr, err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	var (
		buff   = bufio.NewReader(stdout)
		block  = make([]byte, 512)
		offset = int64(0)
	)

	for {
		cnt, err := buff.Read(block)
		if err != nil && err != io.EOF {
			return constant.EmptyStr, err
		}
		if 0 == cnt {
			break
		}
		dumpSql.WriteAt(block[:cnt], offset)
		if err != nil && err != io.EOF {
			return constant.EmptyStr, err
		}
		offset += int64(cnt)
	}
	if err := cmd.Wait(); err != nil {
		return constant.EmptyStr, err
	}
	return dumpSqlPath, nil
}

// genDumpCmdParam
// mysqldump -u <user> -h <example.com> -P <port> -p <database>
func (*DumpTool) genDumpCmdParam(param map[string]any) []string {
	return []string{
		fmt.Sprintf("-u%s", param["user"]),
		fmt.Sprintf("-h%s", param["host"]),
		fmt.Sprintf("-P%s", param["port"]),
		fmt.Sprintf("-p%s", param["password"]),
	}
}
