package guesthandlers

import (
	"context"
	"io/ioutil"
	"net/http"

	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/appsrv"
	"yunion.io/x/onecloud/pkg/hostman/guestman"
	"yunion.io/x/onecloud/pkg/hostman/hostutils"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/util/fileutils2"
)

func guestPrepareImportFormLibvirt(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	_, _, body := appsrv.FetchEnv(ctx, w, r)
	config := &compute.SLibvirtHostConfig{}
	err := body.Unmarshal(config)
	if err != nil {
		hostutils.Response(ctx, w, httperrors.NewInputParameterError("Parse params to libvirt config error %s", err))
		return
	}
	if len(config.XmlFilePath) == 0 {
		hostutils.Response(ctx, w, httperrors.NewMissingParameterError("xml_file_path"))
		return
	}
	if !fileutils2.Exists(config.XmlFilePath) {
		hostutils.Response(ctx, w,
			httperrors.NewBadRequestError("xml_file_path %s not found", config.XmlFilePath))
		return
	}

	if len(config.Servers) == 0 {
		hostutils.Response(ctx, w, httperrors.NewMissingParameterError("servers"))
		return
	}

	if len(config.MonitorPath) > 0 {
		if _, err := ioutil.ReadDir(config.MonitorPath); err != nil {
			hostutils.Response(ctx, w,
				httperrors.NewBadRequestError("Monitor path %s can't open as dir: %s", config.MonitorPath, err))
			return
		}
	}

	hostutils.DelayTask(ctx, guestman.GetGuestManager().PrepareImportFromLibvirt, config)
	hostutils.ResponseOk(ctx, w)
}

func guestCreateFromLibvirt(ctx context.Context, sid string, body jsonutils.JSONObject) (interface{}, error) {
	err := guestman.GetGuestManager().PrepareCreate(sid)
	if err != nil {
		return nil, err
	}

	iGuestDesc, err := body.Get("desc")
	if err != nil {
		return nil, httperrors.NewMissingParameterError("desc")
	}
	guestDesc, ok := iGuestDesc.(*jsonutils.JSONDict)
	if !ok {
		return nil, httperrors.NewInputParameterError("desc is not dict")
	}

	iDisksPath, err := body.Get("disks_path")
	if err != nil {
		return nil, httperrors.NewMissingParameterError("disks_path")
	}
	disksPath, ok := iDisksPath.(*jsonutils.JSONDict)
	if !ok {
		return nil, httperrors.NewInputParameterError("disks_path is not dict")
	}

	monitorPath, _ := body.GetString("monitor_path")
	if len(monitorPath) > 0 && !fileutils2.Exists(monitorPath) {
		return nil, httperrors.NewBadRequestError("Monitor path %s not found", monitorPath)
	}

	hostutils.DelayTask(ctx, guestman.GetGuestManager().GuestCreateFromLibvirt,
		&guestman.SGuestCreateFromLibvirt{sid, monitorPath, guestDesc, disksPath})
	return nil, nil
}

func guestCreateFromEsxi(ctx context.Context, sid string, body jsonutils.JSONObject) (interface{}, error) {
	err := guestman.GetGuestManager().PrepareCreate(sid)
	if err != nil {
		return nil, err
	}

	iGuestDesc, err := body.Get("desc")
	if err != nil {
		return nil, httperrors.NewMissingParameterError("desc")
	}
	guestDesc, ok := iGuestDesc.(*jsonutils.JSONDict)
	if !ok {
		return nil, httperrors.NewInputParameterError("desc is not dict")
	}
	var disksAccessInfo = guestman.SEsxiAccessInfo{}
	err = body.Unmarshal(&disksAccessInfo, "esxi_access_info")
	if err != nil {
		return nil, httperrors.NewMissingParameterError("esxi_access_info")
	}
	hostutils.DelayTask(ctx, guestman.GetGuestManager().GuestCreateFromEsxi,
		&guestman.SGuestCreateFromEsxi{sid, guestDesc, disksAccessInfo})
	return nil, nil
}
