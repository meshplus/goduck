#pier --repo /root/.pier appchain register --appchain-id "ethappchain1" --name "ethereum1" --type "ETH" --trustroot "/root/.pier/ethereum/ether.validators" --broker 0x857133c5C69e6Ce66F7AD46F200B9B3573e77582 --desc "desc" --master-rule "0x00000000000000000000000000000000000000a2" --rule-url "http://github.com" --admin 0x45edbfC20717BF04dd9374aC7bCF434A76e33CEb --reason "reason"
#pier --repo /root/.pier appchain service register --appchain-id ethappchain1 --service-id "0x30c5D3aeb4681af4D13384DBc2a717C51cb1cc11" --name "eth1_transfer" --intro "" --ordered 1 --type CallContract --permit "" --details "test" --reason "reason"
#pier --repo /root/.pier appchain service register --appchain-id ethappchain1 --service-id "0xe95C4c9D9DFeAdC8aD80F87de3F36476DcDdE9F4" --name "eth1_data_swapper" --intro "" --ordered 1 --type CallContract --permit "" --details "test" --reason "reason"
pier --repo /root/.pier start