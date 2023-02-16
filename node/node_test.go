package node

import (
	"crypto/sha256"
	"fmt"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func ReadConfigFile(path string) *config.Config {
	viper.AddConfigPath(filepath.Join(path, "config"))
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	cfg := &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}
	viper.Reset()
	return cfg
}

func AdjustHomePath(cfg *config.Config) {
	cfg.BasicConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.BasicConfig.Home)
	cfg.P2PConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.P2PConfig.Home)
	cfg.ConsensusConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.ConsensusConfig.Home)
	cfg.TxsPoolConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.TxsPoolConfig.Home)
}

func CreateNode(i int) *Node {
	dir := fmt.Sprintf("../node%d", i)
	cfg := ReadConfigFile(dir)
	AdjustHomePath(cfg)
	logger := log.New("node", i)
	logger.SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stdout, log.TerminalFormat(true))))
	log.PrintOrigins(true)

	node, err := NewNode(cfg, logger, DefaultProvider())
	if err != nil {
		panic(err)
	}
	return node
}

func TestCreateAndStartNode(t *testing.T) {
	nodes := make([]*Node, 4)
	nodes[0] = CreateNode(0)
	nodes[1] = CreateNode(1)
	nodes[2] = CreateNode(2)
	nodes[3] = CreateNode(3)

	for i := 0; i < len(nodes); i++ {
		go func(i int) { assert.Nil(t, nodes[i].Start()) }(i)
	}

	time.Sleep(time.Second * 6)

	for i := 0; i < 10; i++ {
		tx := []byte(fmt.Sprintf("%x=%x", []byte("number"), []byte(fmt.Sprintf("%d", i))))
		err := nodes[i%4].txsPool.CheckTx(tx, nodes[i%4].nodeInfo.ID())
		assert.Nil(t, err)
		time.Sleep(time.Millisecond * 1)
	}

	time.Sleep(time.Second * 6)

	fmt.Println("修改前")
	fmt.Println(nodes[0].blockStore.LoadBlockByHeight(2).String())

	time.Sleep(time.Second * 1)

	nodes[0].State().RedactBlock(2, 1, []byte("学校"), []byte("信息工程大学"))

	time.Sleep(time.Second * 3)

	fmt.Println("修改后")
	fmt.Println(nodes[0].blockStore.LoadBlockByHeight(2).String())

	//nodes[0].State().RedactBlock(2, 0, []byte("学校"), []byte("西北工业大学"))
	//
	//time.Sleep(time.Second * 10)
	//
	//fmt.Println("再次修改后")
	//for _, n := range nodes {
	//	fmt.Println(n.blockStore.LoadBlockByHeight(2))
	//}

	//fmt.Println("第三阶段...")
	//for i := 0; i < 3; i++ {
	//	tx := []byte(fmt.Sprintf("number=%d", i+1005))
	//	err := nodes[i%4].txsPool.CheckTx(tx, nodes[i%4].nodeInfo.ID())
	//	assert.Nil(t, err)
	//	time.Sleep(time.Second * 3)
	//}
	//time.Sleep(time.Second * 10)
	//fmt.Println("第四阶段...")
	//for i := 0; i < 2; i++ {
	//	tx := []byte(fmt.Sprintf("number=%d", i+1008))
	//	err := nodes[i%4].txsPool.CheckTx(tx, nodes[i%4].nodeInfo.ID())
	//	assert.Nil(t, err)
	//	time.Sleep(time.Second * 4)
	//}

	select {}
}

func TestName(t *testing.T) {
	s := []byte("1662752367981358478140862759608613144712037411041726893081220576543761682341626337031682502793611742798297090067115192809589292260191992491759811668357940529302152247884394812472019483640159908766027008134032444363930144988172130437679673657583637403155234891485580057349421083124028111804888049368891875025755847956359804119109679138462556349183134351695790497601804060205090351315972506082974845195058458011974008658040900287281558721007107400150104188074180141394662418123318666960689671048548289800563884586402189757218481713197654452771312550848053587721722732290006708751317506037753322744090498429913628376037667925041864917682923438812769566723957312338774319920883186515619329541746744661674476114263639617748568812171787637166786323870426769174676705236399805817315195864829492568131311736026826043924311582407870128363533167047542404720511787733624527839598948652474665639647051527698006350829298562529512323002517106770902258911644298370932129654229431569563566849994343997463834695487028266066122544458464514669824159616114147931001920155950208477368780619486335274037478715737995572028015499457842682225309074478614426350459364409621703813607924006148955595133843493100077504998445298343668317746098695274560197070053489615358608951158091497025380302705001444760101299349419364059419624803150796396608271192017835694608992229147179073325811932434822969748155453991211725781682924997489541937605299817087192452368277042335132448236358185127129306502720532886967462631646193062294367718518351039387483697941072104217946464703793908805216223057785614351986038275745456086211325203768098426260825528368201384939939521354882308108246133649432062715070024801117463567306286974536258926888372532402526048445731940174974953302455303332634948161945727759655238684984479407187310856881043062887096288871862354984886141005410532495884761865552034014013287419236379788622479316727140876196831674249611057821135341362825209900491867537469824917190695989268380153941955783583175562541144468716346225097338997566581239970594151841653650394444660406589638809507290932760295704837424962096674366162062312062986058718799954586217957733229731688650976714326669696357080979273524353715261980220364664845977280878256635958876308985694990378259055823171627481587080416571684574004756258355128818347850983582531805705399739700756381977732512467788720001180888200674468641846564157195349550941245388525490921054291487413422969873297688082820557865529347023754050099047464947406681525848070382115608749501569537539842053228651756001150293607740300613484600474636590253192215215364062316681398998325932612281998310921890694703120863094641099695803090536040625887759707381826990061019778613858475192129810681902364322792714629906022970367593348572720008251707187389611992185027682242013791680917113838635300815572941922899535502400805374743682783414045396100508368074571234497859368858920200174457742494283065870949504795642148661131166045467489358789392424974086847253361563245684463563806186913556157591540542727672227610262224872856715674332449656037536604864184723846755655105419155792000")
	h := sha256.New()
	t.Log(h.Sum(s))
}

// 1662752367981358478140862759608613144712037411041726893081220576543761682341626337031682502793611742798297090067115192809589292260191992491759811668357940529302152247884394812472019483640159908766027008134032444363930144988172130437679673657583637403155234891485580057349421083124028111804888049368891875025755847956359804119109679138462556349183134351695790497601804060205090351315972506082974845195058458011974008658040900287281558721007107400150104188074180141394662418123318666960689671048548289800563884586402189757218481713197654452771312550848053587721722732290006708751317506037753322744090498429913628376037667925041864917682923438812769566723957312338774319920883186515619329541746744661674476114263639617748568812171787637166786323870426769174676705236399805817315195864829492568131311736026826043924311582407870128363533167047542404720511787733624527839598948652474665639647051527698006350829298562529512323002517106770902258911644298370932129654229431569563566849994343997463834695487028266066122544458464514669824159616114147931001920155950208477368780619486335274037478715737995572028015499457842682225309074478614426350459364409621703813607924006148955595133843493100077504998445298343668317746098695274560197070053489615358608951158091497025380302705001444760101299349419364059419624803150796396608271192017835694608992229147179073325811932434822969748155453991211725781682924997489541937605299817087192452368277042335132448236358185127129306502720532886967462631646193062294367718518351039387483697941072104217946464703793908805216223057785614351986038275745456086211325203768098426260825528368201384939939521354882308108246133649432062715070024801117463567306286974536258926888372532402526048445731940174974953302455303332634948161945727759655238684984479407187310856881043062887096288871862354984886141005410532495884761865552034014013287419236379788622479316727140876196831674249611057821135341362825209900491867537469824917190695989268380153941955783583175562541144468716346225097338997566581239970594151841653650394444660406589638809507290932760295704837424962096674366162062312062986058718799954586217957733229731688650976714326669696357080979273524353715261980220364664845977280878256635958876308985694990378259055823171627481587080416571684574004756258355128818347850983582531805705399739700756381977732512467788720001180888200674468641846564157195349550941245388525490921054291487413422969873297688082820557865529347023754050099047464947406681525848070382115608749501569537539842053228651756001150293607740300613484600474636590253192215215364062316681398998325932612281998310921890694703120863094641099695803090536040625887759707381826990061019778613858475192129810681902364322792714629906022970367593348572720008251707187389611992185027682242013791680917113838635300815572941922899535502400805374743682783414045396100508368074571234497859368858920200174457742494283065870949504795642148661131166045467489358789392424974086847253361563245684463563806186913556157591540542727672227610262224872856715674332449656037536604864184723846755655105419155792000
