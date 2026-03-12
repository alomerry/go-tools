package notify

const (
  NotifyTypeAlarm  NotifyType = "alarm"
  NotifyTypeSystem NotifyType = "system"
)

const (
  NotifyGroupInfo   NotifyGroup = "info"
  NotifyGroupAlarm  NotifyGroup = "alarm"
  NotifyGroupAlert  NotifyGroup = "alert"
  NotifyGroupBuild  NotifyGroup = "build"
  NotifyGroupDeploy NotifyGroup = "deploy"
)

const (
  NotifySenderKook    NotifySenderType = "kook"
  NotifySenderBark    NotifySenderType = "bark"
  NotifySenderConsole NotifySenderType = "console"
)

const (
  NotifyMsgExtGroup = "group"
  NotifyMsgExtLevel = "level"
)

var (
  types = []NotifyType{
    NotifyTypeAlarm,
    NotifyTypeSystem,
  }
  
  senderTypes = []NotifySenderType{
    NotifySenderKook,
    NotifySenderBark,
    NotifySenderConsole,
  }
  
  groups = []NotifyGroup{
    NotifyGroupAlarm,
    NotifyGroupAlert,
    NotifyGroupBuild,
    NotifyGroupDeploy,
    NotifyGroupInfo,
  }
)

type NotifySenderType string
type NotifyGroup string
type NotifyType string

func ToNotifyType(str string) NotifyType {
  for i := range types {
    if string(types[i]) == str {
      return types[i]
    }
  }
  
  return ""
}

func ToNotifySenderType(str string) NotifySenderType {
  for i := range senderTypes {
    if string(senderTypes[i]) == str {
      return senderTypes[i]
    }
  }
  
  return ""
}

func (n NotifyType) String() string {
  return string(n)
}

func (n NotifyGroup) String() string {
  return string(n)
}

func ToNotifyGroup(str string) NotifyGroup {
  for i := range groups {
    if string(groups[i]) == str {
      return groups[i]
    }
  }
  return NotifyGroupInfo
}

