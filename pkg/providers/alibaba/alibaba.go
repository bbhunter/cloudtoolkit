package alibaba

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	_bss "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/bss"
	_dns "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/dns"
	_ecs "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/ecs"
	_oss "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/oss"
	_ram "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/ram"
	_rds "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/rds"
	_sas "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/sas"
	_sms "github.com/404tk/cloudtoolkit/pkg/providers/alibaba/sms"
	"github.com/404tk/cloudtoolkit/pkg/schema"
	"github.com/404tk/cloudtoolkit/utils"
	"github.com/404tk/cloudtoolkit/utils/cache"
	"github.com/404tk/cloudtoolkit/utils/table"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
)

// Provider is a data provider for alibaba API
type Provider struct {
	vendor string
	cred   *credentials.StsTokenCredential
	region string
}

// New creates a new provider client for alibaba API
func New(options schema.Options) (*Provider, error) {
	accessKey, ok := options.GetMetadata(utils.AccessKey)
	if !ok {
		return nil, &schema.ErrNoSuchKey{Name: utils.AccessKey}
	}
	secretKey, ok := options.GetMetadata(utils.SecretKey)
	if !ok {
		return nil, &schema.ErrNoSuchKey{Name: utils.SecretKey}
	}
	region, _ := options.GetMetadata(utils.Region)
	token, _ := options.GetMetadata(utils.SecurityToken)
	cred := credentials.NewStsTokenCredential(accessKey, secretKey, token)

	// Get current username
	stsclient, err := sts.NewClientWithOptions("cn-hangzhou", sdk.NewConfig(), cred)
	request := sts.CreateGetCallerIdentityRequest()
	request.Scheme = "https"
	response, err := stsclient.GetCallerIdentity(request)
	if err != nil {
		return nil, err
	}
	accountArn := response.Arn
	var userName string
	if len(accountArn) >= 4 && accountArn[len(accountArn)-4:] == "root" {
		userName = "root"
	} else {
		if u := strings.Split(accountArn, "/"); len(u) > 1 {
			userName = u[1]
		}
	}
	msg := "[+] Current user: " + userName
	cache.Cfg.CredInsert(userName, options)
	log.Printf(msg)

	return &Provider{
		vendor: "alibaba",
		cred:   cred,
		region: region,
	}, nil
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return p.vendor
}

// Resources returns the provider for a resource deployment source.
func (p *Provider) Resources(ctx context.Context) (schema.Resources, error) {
	list := schema.NewResources()
	list.Provider = p.vendor
	var err error
	for _, product := range utils.Cloudlist {
		switch product {
		case "balance":
			d := &_bss.Driver{Cred: p.cred, Region: p.region}
			d.QueryAccountBalance(ctx)
		case "host":
			ecsprovider := &_ecs.Driver{Cred: p.cred, Region: p.region}
			list.Hosts, err = ecsprovider.GetResource(ctx)
		case "domain":
			dnsprovider := &_dns.Driver{Cred: p.cred, Region: p.region}
			list.Domains, err = dnsprovider.GetDomains(ctx)
		case "account":
			ramprovider := &_ram.Driver{Cred: p.cred, Region: p.region}
			list.Users, err = ramprovider.GetRamUser(ctx)
		case "database":
			rdsprovider := &_rds.Driver{Cred: p.cred, Region: p.region}
			list.Databases, err = rdsprovider.GetDatabases(ctx)
		case "bucket":
			ossprovider := &_oss.Driver{Cred: p.cred, Region: p.region}
			list.Storages, err = ossprovider.GetBuckets(ctx)
		case "sms":
			smsprovider := &_sms.Driver{Cred: p.cred, Region: p.region}
			list.Sms, err = smsprovider.GetResource(ctx)
		default:
		}
	}

	return list, err
}

func (p *Provider) UserManagement(action, args_1, args_2 string) {
	r := &_ram.Driver{Cred: p.cred, Region: p.region}
	switch action {
	case "add":
		r.UserName = args_1
		r.PassWord = args_2
		r.AddUser()
	case "del":
		r.UserName = args_1
		r.DelUser()
	case "shadow":
		r.RoleName = args_1
		r.AccountId = args_2
		r.AddRole()
	case "delrole":
		r.RoleName = args_1
		r.DelRole()
	default:
		log.Println("[-] Please set metadata like \"add username password\" or \"del username\"")
	}
}

func (p *Provider) BucketDump(ctx context.Context, action, bucketname string) {
	ossdrvier := &_oss.Driver{Cred: p.cred, Region: p.region}
	switch action {
	case "list":
		var infos = make(map[string]string)
		if bucketname == "all" {
			buckets, _ := ossdrvier.GetBuckets(context.Background())
			for _, b := range buckets {
				infos[b.BucketName] = b.Region
			}
		} else {
			infos[bucketname] = p.region
		}
		ossdrvier.ListObjects(ctx, infos)
	case "total":
		var infos = make(map[string]string)
		if bucketname == "all" {
			buckets, _ := ossdrvier.GetBuckets(context.Background())
			for _, b := range buckets {
				infos[b.BucketName] = b.Region
			}
		} else {
			infos[bucketname] = p.region
		}
		ossdrvier.TotalObjects(ctx, infos)
	default:
		log.Println("[-] `list all` or `total all`.")
	}
}

func (p *Provider) EventDump(action, sourceIp string) {
	d := _sas.Driver{Cred: p.cred}
	switch action {
	case "dump":
		events, err := d.DumpEvents()
		if err != nil {
			log.Println("[-]", err)
			return
		}
		if len(events) == 0 {
			return
		}
		table.Output(events)
		if utils.DoSave {
			filename := time.Now().Format("20060102150405.log")
			path := fmt.Sprintf("%s/%s_eventdump_%s", utils.LogDir, p.Name(), filename)
			table.FileOutput(path, events)
			log.Printf("[+] Output written to [%s]\n", path)
		}
	case "whitelist":
		d.HandleEvents(sourceIp) // sourceIp here means SecurityEventIds
	default:
		log.Println("[-] Please set metadata like \"dump all\"")
	}
}
