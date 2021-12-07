// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

import "github.com/ethereum/go-ethereum/common"

var MainnetMasternodes = []string{
	//"0x7b57ac80bc6722050f62f132e7c805a549b4d929",
	//"0xf6913117dff6de6836d5ce433cf5ebe2e09b3a8e",
	//"0x6ea9a73f144d7dde8642a91f911441bc886804ef",
	//"0x87a1c02d0c3bc1a8f8d75554f0b9d71c4f3ff1c3",
	//"0xd89af0b36f5be8b485f6eb6a35dcb0fe3e5fda8b",
	//"0xe97b42706764b6d106b85209ea864df208250497",
	//"0xd5aabb2df7a7f7abc58f7bbd89d73899b2123e4d",
	//"0x0f5d04675b75727c8bf70461e30b5858de30617e",
	//"0x969ed10bbb4d63379ef1598abe2be33ebd9dfdda",
	//"0xc6f25bbafc27f99dbd9fbd762f6a4232788e070c",
	//"0xa78fb47cde509cf584861a831053376fe33a41d0",
	//"0x156e0a9a702f6c800638302e9e4b1300e687f033",
	//"0x965f4b6ce4f2e7258aeeebb0fa07047e313b23d2",
	//"0x9a2420d0d633f49b366a824a16173b2e48b0293e",
	//"0xb12144b2a128d47fa3f4c745ea389831b6196f8b",
	//"0x057a221e562ebd97f2ab417b2e997bef685397f9",
	//"0xc5bb7cc0aec398de9659242b7daa10f9abc817be",
	//"0x0e91ca2d6c1013af039fc9378d731e1e975ac8a3",
	//"0xde0778dfbd67a9ce8294b15aa07911e7a765bde5",
	//"0x12f986390eb6be0ef1808f2ffa1ea8af39c86a86",
	//"0xeb656cfa7ace74002b09a2110586703aaf25d8c9",

	//"0xe344322fbf7b9efc7d0d023ecc8d675458a4bfbe",
	//"0x5c1cd70030db96c11e0fb9e45f11e29a1b6e0950",
	//"0x83423eeb31bbe9bbc5b9123948ef40cc2e6e7a5a",
	//"0xf3396872f78e2007de7edbcebdf18cfcb62358b4",
	//"0xbe80e1fbb7ef09743adcf1ff5ffe0601d3d0ae16",
	//"0xe718711d5d676bd82e0bb27da036f0e9198248a7",
	//"0x9f3ead0c22468b62094d9efc524269cd2334b714",
	//"0xeaa7db59ca62a48e2b847df8066e050da1677268",
	//"0xb8fcdb521a168d330e6673851b6d73c8c55cf505",
	//"0xb77b5ecbd3dd1913b2b89b56b0fe468e6a08af2e",
	//"0xe86d216e45e40f10e70b4b9ecd0994ef765466c9",
	//"0xf4d35a88c39bc7de0b44811a80ec94f135e6c647",
	//"0x056e1026fa6ce9e301fc6b804c0674e2aabe28a0",
	//"0xb8391fadbcb12d7b7d17d59d89785d8c9557736c",
	//"0xc23911940ecb1a883abfbec91237fca9d37eb84b",
	//"0x49017899cbf5ae1660fcab006e11ba81c92b28e0",
	//"0x8bab5a3a5d15530a2402a88dc235420cc5e1670b",
	//"0x1f7ebb7bf89f32d585745d1c81029645740e945d",
	//"0x88f10a286df74223aff5cc234fc3c985815a9d13",
	//"0x2e100312f7a430872de0210dc267f1bb50de22df",
	//"0x0d90732b68c7da1fa47e3264cce4bffa968c373b",

	"0x742e63e869063f8821d3f0d91a61565579025f5b",
	"0xeb599e9bfb3e8941f88580c1de01eab2a4ca5f35",
	"0xa88603c2963d7d5ebbb2652a9aa2541638bde196",
	"0x7cb4e17f37a62b57d27a86f24c9eabd72901abeb",
	"0x4bfe994ea05202bcfe60c5bc359d1c473268ae2e",
	"0xe283268f88f722d67ec8e8b2cabd2c990710156f",
	"0xb0ccb449a61bb0a17041fc7057f645019d34e045",
	"0x440c840609689765acd78756a7a2717df47bc8b5",
	"0xe88288cc5e00fd5b7ba73e29bcde6dc5f6b77f18",
	"0x61ceedd74303c5df4aca68326475548f7fdae009",
	"0x8d0a8f6dabde76683caa4d68092ea061eca0d51e",
	"0xf36052546b6e491edac0bb3f1c1859e73ff63df1",
	"0xd985a6f0d0677f21c14658fd59c1ab1a2dc06d5b",
	"0xb373ea7d4b5e98899f74aca5da79b298751a4c3a",
	"0xe105ba06547a43dac5b6c0ff0967874f3f903ad4",
	"0x2ae3f67688ab3847f78d31bf58d7bb073206fbf4",
	"0x0b696eb75b1eec395f6a85818aa940d96a47046b",
	"0x429718401db69a877f68a25021261da3c714118f",
	"0x098065c285ddf5f376a317ea3af25c3bb8d14bc4",
	"0xe1e2665893e3d2791fb4d152214e71a22d5373d8",
	"0xda406640b32cc19b0b4efed4010a8b3fa196d539",
}

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Ethereum network.
var MainnetBootnodes = []string{
	// Ethereum Foundation Go Bootnodes
	//"enode://4864112705c40f3750cdad518cb2941a2256c1833ae902acf65652e65035e598d5196f5ed1087b0849cdc58bf8404f4bd36e1779aacfa4da538ce159ad31cdbb@154.85.60.46:37303",
	"enode://4a391531a8e3f35ef77af1eb8441aeeaa99c3a4e6eafe27f8f1ab259ac7b9f9cb2ee065082ec45ad4edf254c763c0354e1175407311cf47a6be9406f95d037cd@101.32.177.140:37303",
}

// RopstenBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var RopstenBootnodes = []string{
	"enode://30b7ab30a01c124a6cceca36863ece12c4f5fa68e3ba9b0b51407ccc002eeed3b3102d20a88f1c1d3c3154e2449317b8ef95090e77b312d5cc39354f86d5d606@52.176.7.10:30303",    // US-Azure geth
	"enode://865a63255b3bb68023b6bffd5095118fcc13e79dcf014fe4e47e065c350c7cc72af2e53eff895f11ba1bbb6a2b33271c1116ee870f266618eadfc2e78aa7349c@52.176.100.77:30303",  // US-Azure parity
	"enode://6332792c4a00e3e4ee0926ed89e0d27ef985424d97b6a45bf0f23e51f0dcb5e66b875777506458aea7af6f9e4ffb69f43f3778ee73c81ed9d34c51c4b16b0b0f@52.232.243.152:30303", // Parity
	"enode://94c15d1b9e2fe7ce56e458b9a3b672ef11894ddedd0c6f247e0f1d3487f52b66208fb4aeb8179fce6e3a749ea93ed147c37976d67af557508d199d9594c35f09@192.81.208.223:30303", // @gpip
}

// SepoliaBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Sepolia test network.
var SepoliaBootnodes = []string{
	// geth
	"enode://9246d00bc8fd1742e5ad2428b80fc4dc45d786283e05ef6edbd9002cbc335d40998444732fbe921cb88e1d2c73d1b1de53bae6a2237996e9bfe14f871baf7066@18.168.182.86:30303",
	// besu
	"enode://ec66ddcf1a974950bd4c782789a7e04f8aa7110a72569b6e65fcd51e937e74eed303b1ea734e4d19cfaec9fbff9b6ee65bf31dcb50ba79acce9dd63a6aca61c7@52.14.151.177:30303",
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
	"enode://a24ac7c5484ef4ed0c5eb2d36620ba4e4aa13b8c84684e1b4aab0cebea2ae45cb4d375b77eab56516d34bfbd3c1a833fc51296ff084b770b94fb9028c4d25ccf@52.169.42.101:30303", // IE
	"enode://343149e4feefa15d882d9fe4ac7d88f885bd05ebb735e547f12e12080a9fa07c8014ca6fd7f373123488102fe5e34111f8509cf0b7de3f5b44339c9f25e87cb8@52.3.158.184:30303",  // INFURA
	"enode://b6b28890b006743680c52e64e0d16db57f28124885595fa03a562be1d2bf0f3a1da297d56b13da25fb992888fd556d4c1a27b1f39d531bde7de1921c90061cc6@159.89.28.211:30303", // AKASHA
}

// GoerliBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Görli test network.
var GoerliBootnodes = []string{
	// Upstream bootnodes
	"enode://011f758e6552d105183b1761c5e2dea0111bc20fd5f6422bc7f91e0fabbec9a6595caf6239b37feb773dddd3f87240d99d859431891e4a642cf2a0a9e6cbb98a@51.141.78.53:30303",
	"enode://176b9417f511d05b6b2cf3e34b756cf0a7096b3094572a8f6ef4cdcb9d1f9d00683bf0f83347eebdf3b81c3521c2332086d9592802230bf528eaf606a1d9677b@13.93.54.137:30303",
	"enode://46add44b9f13965f7b9875ac6b85f016f341012d84f975377573800a863526f4da19ae2c620ec73d11591fa9510e992ecc03ad0751f53cc02f7c7ed6d55c7291@94.237.54.114:30313",
	"enode://b5948a2d3e9d486c4d75bf32713221c2bd6cf86463302339299bd227dc2e276cd5a1c7ca4f43a0e9122fe9af884efed563bd2a1fd28661f3b5f5ad7bf1de5949@18.218.250.66:30303",

	// Ethereum Foundation bootnode
	"enode://a61215641fb8714a373c80edbfa0ea8878243193f57c96eeb44d0bc019ef295abd4e044fd619bfc4c59731a73fb79afe84e9ab6da0c743ceb479cbb6d263fa91@3.11.147.67:30303",

	// Goerli Initiative bootnodes
	"enode://a869b02cec167211fb4815a82941db2e7ed2936fd90e78619c53eb17753fcf0207463e3419c264e2a1dd8786de0df7e68cf99571ab8aeb7c4e51367ef186b1dd@51.15.116.226:30303",
	"enode://807b37ee4816ecf407e9112224494b74dd5933625f655962d892f2f0f02d7fbbb3e2a94cf87a96609526f30c998fd71e93e2f53015c558ffc8b03eceaf30ee33@51.15.119.157:30303",
	"enode://a59e33ccd2b3e52d578f1fbd70c6f9babda2650f0760d6ff3b37742fdcdfdb3defba5d56d315b40c46b70198c7621e63ffa3f987389c7118634b0fefbbdfa7fd@51.15.119.157:40303",
}

var V5Bootnodes = []string{
	// Teku team's bootnode
	"enr:-KG4QOtcP9X1FbIMOe17QNMKqDxCpm14jcX5tiOE4_TyMrFqbmhPZHK_ZPG2Gxb1GE2xdtodOfx9-cgvNtxnRyHEmC0ghGV0aDKQ9aX9QgAAAAD__________4JpZIJ2NIJpcIQDE8KdiXNlY3AyNTZrMaEDhpehBDbZjM_L9ek699Y7vhUJ-eAdMyQW_Fil522Y0fODdGNwgiMog3VkcIIjKA",
	"enr:-KG4QDyytgmE4f7AnvW-ZaUOIi9i79qX4JwjRAiXBZCU65wOfBu-3Nb5I7b_Rmg3KCOcZM_C3y5pg7EBU5XGrcLTduQEhGV0aDKQ9aX9QgAAAAD__________4JpZIJ2NIJpcIQ2_DUbiXNlY3AyNTZrMaEDKnz_-ps3UUOfHWVYaskI5kWYO_vtYMGYCQRAR3gHDouDdGNwgiMog3VkcIIjKA",
	// Prylab team's bootnodes
	"enr:-Ku4QImhMc1z8yCiNJ1TyUxdcfNucje3BGwEHzodEZUan8PherEo4sF7pPHPSIB1NNuSg5fZy7qFsjmUKs2ea1Whi0EBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQOVphkDqal4QzPMksc5wnpuC3gvSC8AfbFOnZY_On34wIN1ZHCCIyg",
	"enr:-Ku4QP2xDnEtUXIjzJ_DhlCRN9SN99RYQPJL92TMlSv7U5C1YnYLjwOQHgZIUXw6c-BvRg2Yc2QsZxxoS_pPRVe0yK8Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQMeFF5GrS7UZpAH2Ly84aLK-TyvH-dRo0JM1i8yygH50YN1ZHCCJxA",
	"enr:-Ku4QPp9z1W4tAO8Ber_NQierYaOStqhDqQdOPY3bB3jDgkjcbk6YrEnVYIiCBbTxuar3CzS528d2iE7TdJsrL-dEKoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhBLf22SJc2VjcDI1NmsxoQMw5fqqkw2hHC4F5HZZDPsNmPdB1Gi8JPQK7pRc9XHh-oN1ZHCCKvg",
	// Lighthouse team's bootnodes
	"enr:-IS4QLkKqDMy_ExrpOEWa59NiClemOnor-krjp4qoeZwIw2QduPC-q7Kz4u1IOWf3DDbdxqQIgC4fejavBOuUPy-HE4BgmlkgnY0gmlwhCLzAHqJc2VjcDI1NmsxoQLQSJfEAHZApkm5edTCZ_4qps_1k_ub2CxHFxi-gr2JMIN1ZHCCIyg",
	"enr:-IS4QDAyibHCzYZmIYZCjXwU9BqpotWmv2BsFlIq1V31BwDDMJPFEbox1ijT5c2Ou3kvieOKejxuaCqIcjxBjJ_3j_cBgmlkgnY0gmlwhAMaHiCJc2VjcDI1NmsxoQJIdpj_foZ02MXz4It8xKD7yUHTBx7lVFn3oeRP21KRV4N1ZHCCIyg",
	// EF bootnodes
	"enr:-Ku4QHqVeJ8PPICcWk1vSn_XcSkjOkNiTg6Fmii5j6vUQgvzMc9L1goFnLKgXqBJspJjIsB91LTOleFmyWWrFVATGngBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhAMRHkWJc2VjcDI1NmsxoQKLVXFOhp2uX6jeT0DvvDpPcU8FWMjQdR4wMuORMhpX24N1ZHCCIyg",
	"enr:-Ku4QG-2_Md3sZIAUebGYT6g0SMskIml77l6yR-M_JXc-UdNHCmHQeOiMLbylPejyJsdAPsTHJyjJB2sYGDLe0dn8uYBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhBLY-NyJc2VjcDI1NmsxoQORcM6e19T1T9gi7jxEZjk_sjVLGFscUNqAY9obgZaxbIN1ZHCCIyg",
	"enr:-Ku4QPn5eVhcoF1opaFEvg1b6JNFD2rqVkHQ8HApOKK61OIcIXD127bKWgAtbwI7pnxx6cDyk_nI88TrZKQaGMZj0q0Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhDayLMaJc2VjcDI1NmsxoQK2sBOLGcUb4AwuYzFuAVCaNHA-dy24UuEKkeFNgCVCsIN1ZHCCIyg",
	"enr:-Ku4QEWzdnVtXc2Q0ZVigfCGggOVB2Vc1ZCPEc6j21NIFLODSJbvNaef1g4PxhPwl_3kax86YPheFUSLXPRs98vvYsoBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhDZBrP2Jc2VjcDI1NmsxoQM6jr8Rb1ktLEsVcKAPa08wCsKUmvoQ8khiOl_SLozf9IN1ZHCCIyg",
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	var net string
	switch genesis {
	case MainnetGenesisHash:
		net = "mainnet"
	case RopstenGenesisHash:
		net = "ropsten"
	case RinkebyGenesisHash:
		net = "rinkeby"
	case GoerliGenesisHash:
		net = "goerli"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".ethdisco.net"
}
