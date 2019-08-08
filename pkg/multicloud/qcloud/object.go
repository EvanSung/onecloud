// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qcloud

import (
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"

	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"yunion.io/x/onecloud/pkg/cloudprovider"
)

type SObject struct {
	bucket *SBucket

	cloudprovider.SBaseCloudObject
}

func (o *SObject) GetIBucket() cloudprovider.ICloudBucket {
	return o.bucket
}

func (o *SObject) GetAcl() cloudprovider.TBucketACLType {
	acl := cloudprovider.ACLDefault
	coscli, err := o.bucket.region.GetCosClient(o.bucket)
	if err != nil {
		log.Errorf("o.bucket.region.GetOssClient error %s", err)
		return acl
	}
	result, _, err := coscli.Object.GetACL(context.Background(), o.Key)
	if err != nil {
		log.Errorf("coscli.Object.GetACL error %s", err)
		return acl
	}
	return cosAcl2CannedAcl(result.AccessControlList)
}

func (o *SObject) SetAcl(aclStr cloudprovider.TBucketACLType) error {
	coscli, err := o.bucket.region.GetCosClient(o.bucket)
	if err != nil {
		return errors.Wrap(err, "o.bucket.region.GetCosClient")
	}
	opts := &cos.ObjectPutACLOptions{}
	opts.Header.XCosACL = string(aclStr)
	_, err = coscli.Object.PutACL(context.Background(), o.Key, opts)
	if err != nil {
		return errors.Wrap(err, "coscli.Object.PutACL")
	}
	return nil
}