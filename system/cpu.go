package system

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

//CPU型号
func CpuModel(args string) string {
	return ExecOutput("cat " + CPUFile + " | grep name | cut -f2 -d: | uniq")
}

//逻辑CPU个数
func CpuNum(args string) string {
	return ExecOutput("cat " + CPUFile + " | grep processor | wc -l")
}

type Cpu struct {
	User                uint64  //从系统启动开始累计到当前时刻，用户态的CPU时间（单位：jiffies） ，不包含 nice值为负进程。1jiffies=0.01秒
	Nice                uint64  //从系统启动开始累计到当前时刻，nice值为负的进程所占用的CPU时间（单位：jiffies）
	System              uint64  //从系统启动开始累计到当前时刻，内核态时间（单位：jiffies）
	Idle                uint64  //从系统启动开始累计到当前时刻，除硬盘IO等待时间以外其它等待时间（单位：jiffies)
	Iowait              uint64  //从系统启动开始累计到当前时刻，硬盘IO等待时间（单位：jiffies）
	Irq                 uint64  //从系统启动开始累计到当前时刻，硬中断时间（单位：jiffies）
	SoftIrq             uint64  //从系统启动开始累计到当前时刻，软中断时间（单位：jiffies）
	Total               uint64  //user + nice + system + idle + iowait
	IoWaitRate          float64 //io等待时间百分比
	SystemRate          float64 //内核态时间百分比
	UserRate            float64 //用户态时间百分比
	IdleRate            float64 //空闲时间百分比
	ProcsBlocked        uint64  //阻塞进程数
	ProcsRunning        uint64  //运行进程数
	IdleRateSum10       float64 //空闲时间百分比10分钟累加和
	IdleRateSum10Times  int     //空闲时间百分比10分钟累加次数
	IdleRate10          float64 //空闲时间10分钟环比
	IdleRate10Last      int64
	IdleRateSum60       float64 //空闲时间百分比60分钟累加和
	IdleRateSum60Times  int     //空闲时间百分比60分钟累加次数
	IdleRate60          float64 //空闲时间60分钟环比
	IdleRate60Last      int64
	IdleRateSumDay      float64 //空闲时间百分比24h累加和
	IdleRateSumDayTimes int     //空闲时间百分比24h累加次数
	IdleRateDay         float64 //空闲时间日同比
	IdleRateDayLast     int64
}

func (this *Cpu) Dump() {
	fmt.Printf("User:%d, Nice:%d, System:%d, Idle:%d, Iowait:%d, Irq:%d, SoftIrq:%d, Total:%d, IoWaitRate:%f\n",
		this.User,
		this.Nice,
		this.System,
		this.Idle,
		this.Iowait,
		this.Irq,
		this.SoftIrq,
		this.Total,
		this.IoWaitRate)
}

func (this *Cpu) Collect() error {
	f, err := os.Open(StatFile)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	readCpuTag := false
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.Contains(line, "cpu") {
			if readCpuTag {
				//读取过，跳过
				continue
			}
			//文件中顺序依次为user, nice, system, idle, iowait, irq, softirq, stealstolen, guest
			strList := strings.Fields(line)
			user, _ := strconv.ParseUint(strList[1], 10, 64)
			nice, _ := strconv.ParseUint(strList[2], 10, 64)
			system, _ := strconv.ParseUint(strList[3], 10, 64)
			idle, _ := strconv.ParseUint(strList[4], 10, 64)
			iowait, _ := strconv.ParseUint(strList[5], 10, 64)
			irq, _ := strconv.ParseUint(strList[6], 10, 64)
			softIrq, _ := strconv.ParseUint(strList[7], 10, 64)
			total := user + nice + system + idle + iowait

			diffTotal := float64(total - this.Total)

			if this.Total <= 0 {
				//第一次采集，没产生时间差，不计算
			} else {
				if diffTotal > 0 {
					//io等待时间百分比
					diffIo := iowait - this.Iowait
					this.IoWaitRate = float64(diffIo) / diffTotal * 100
					//内核态时间百分比
					diffSystem := system - this.System
					this.SystemRate = float64(diffSystem) / diffTotal * 100
					//用户态时间百分比
					diffUser := user - this.User
					this.UserRate = float64(diffUser) / diffTotal * 100
					//空闲时间百分比
					this.IdleRate = float64(idle-this.Idle) / diffTotal * 100
				}
			}

			this.User = user
			this.Nice = nice
			this.System = system
			this.Idle = idle
			this.Iowait = iowait
			this.Irq = irq
			this.SoftIrq = softIrq
			this.Total = total
			readCpuTag = true

		} else if strings.Contains(line, "procs_blocked") {
			strList := strings.Fields(line)
			this.ProcsBlocked, _ = strconv.ParseUint(strList[1], 10, 64)
		} else if strings.Contains(line, "procs_running") {
			strList := strings.Fields(line)
			this.ProcsRunning, _ = strconv.ParseUint(strList[1], 10, 64)
		} else {
			continue
		}
	}
	return nil
}

//io等待时间百分比
func (this *Cpu) IoWaitRateFunc(args string) string {
	return FloatToString(this.IoWaitRate)
}

//内核态时间百分比
func (this *Cpu) SystemRateFunc(args string) string {
	return FloatToString(this.SystemRate)
}

//用户态时间百分比
func (this *Cpu) UserRateFunc(args string) string {
	return FloatToString(this.UserRate)
}

//空闲时间百分比
func (this *Cpu) IdleRateFunc(args string) string {
	//this.AddIdleRate(this.IdleRate)
	return FloatToString(this.IdleRate)
}

//阻塞进程数
func (this *Cpu) ProcsBlockedFunc(args string) string {
	return strconv.FormatUint(this.ProcsBlocked, 10)
}

//运行进程数
func (this *Cpu) ProcsRunningFunc(args string) string {
	return strconv.FormatUint(this.ProcsRunning, 10)
}

//返回某个进程的cpu使用率
func CpuUsedRateByProc(proc string) string {
	//return ExecOutput("ps axuwww|grep " + proc + "|grep -v grep|awk 'BEGIN{sum=0}{sum+=$3 }END{print sum}'")
	return ExecOutput("top -n 1 -b |grep '" + proc + "'|grep -v grep|awk 'BEGIN{sum=0}{sum+=$9 }END{print sum}'")
}
