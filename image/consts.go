// Constants

package image

const SWAP_LABEL = "SWAP"
const SLASH = "/"

const (
	ACTION_CREATE  = "create"
	ACTION_UPLOAD  = "upload"
	ACTION_REMOVE  = "remove"
	ACTION_START   = "start"
	ACTION_STOP    = "stop"
	ACTION_RESTART = "restart"
	ACTION_RELOAD  = "rload"
	ACTION_APPEND  = "append"
	ACTION_REPLACE = "replace"
	ACTION_INSTALL = "install"
)

const (
	SVC_STATUS_ON  = "on"
	SVC_STATUS_OFF = "off"
)

const (
	SVC_TYPE_SYSV    = "sysv"
	SVC_TYPE_UPSTART = "upstart"
)

const (
	PKG_TYPE_RPM = "rpm"
	PKG_TYPE_DEB = "deb"
)

const (
	ITEM_TYPE_FILE = "file"
	ITEM_TYPE_DIR  = "directory"
	ITEM_TYPE_LINK = "link"
)

const (
	PRE_SCRIPTS = iota
	POST_SCRIPTS
)

const (
	INJ_TYPE_ALTEON_CONFIG = iota
	INJ_TYPE_FILE
	INJ_TYPE_DIR
	INJ_TYPE_QUIT
)
