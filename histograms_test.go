package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/prometheus/common/expfmt"
)

func TestHistogramConversion(t *testing.T) {

	input := `
# HELP raft_process_logcommit_latency Latency histogram for committing Raft log entries
# TYPE raft_process_logcommit_latency histogram
raft_process_logcommit_latency_bucket{store="1",le="69631"} 1
raft_process_logcommit_latency_bucket{store="1",le="73727"} 71
raft_process_logcommit_latency_bucket{store="1",le="77823"} 791
raft_process_logcommit_latency_bucket{store="1",le="81919"} 4030
raft_process_logcommit_latency_bucket{store="1",le="86015"} 12930
raft_process_logcommit_latency_bucket{store="1",le="90111"} 33860
raft_process_logcommit_latency_bucket{store="1",le="94207"} 75094
raft_process_logcommit_latency_bucket{store="1",le="98303"} 147041
raft_process_logcommit_latency_bucket{store="1",le="102399"} 262557
raft_process_logcommit_latency_bucket{store="1",le="106495"} 434051
raft_process_logcommit_latency_bucket{store="1",le="110591"} 667155
raft_process_logcommit_latency_bucket{store="1",le="114687"} 963529
raft_process_logcommit_latency_bucket{store="1",le="118783"} 1.32395e+06
raft_process_logcommit_latency_bucket{store="1",le="122879"} 1.749268e+06
raft_process_logcommit_latency_bucket{store="1",le="126975"} 2.234776e+06
raft_process_logcommit_latency_bucket{store="1",le="131071"} 2.779248e+06
raft_process_logcommit_latency_bucket{store="1",le="139263"} 4.027729e+06
raft_process_logcommit_latency_bucket{store="1",le="147455"} 5.471341e+06
raft_process_logcommit_latency_bucket{store="1",le="155647"} 7.076698e+06
raft_process_logcommit_latency_bucket{store="1",le="163839"} 8.791206e+06
raft_process_logcommit_latency_bucket{store="1",le="172031"} 1.0565046e+07
raft_process_logcommit_latency_bucket{store="1",le="180223"} 1.2347269e+07
raft_process_logcommit_latency_bucket{store="1",le="188415"} 1.4093267e+07
raft_process_logcommit_latency_bucket{store="1",le="196607"} 1.5769488e+07
raft_process_logcommit_latency_bucket{store="1",le="204799"} 1.735897e+07
raft_process_logcommit_latency_bucket{store="1",le="212991"} 1.884948e+07
raft_process_logcommit_latency_bucket{store="1",le="221183"} 2.0239109e+07
raft_process_logcommit_latency_bucket{store="1",le="229375"} 2.152608e+07
raft_process_logcommit_latency_bucket{store="1",le="237567"} 2.2715712e+07
raft_process_logcommit_latency_bucket{store="1",le="245759"} 2.3812696e+07
raft_process_logcommit_latency_bucket{store="1",le="253951"} 2.4821616e+07
raft_process_logcommit_latency_bucket{store="1",le="262143"} 2.5748801e+07
raft_process_logcommit_latency_bucket{store="1",le="278527"} 2.7391408e+07
raft_process_logcommit_latency_bucket{store="1",le="294911"} 2.8793933e+07
raft_process_logcommit_latency_bucket{store="1",le="311295"} 2.9999262e+07
raft_process_logcommit_latency_bucket{store="1",le="327679"} 3.1046358e+07
raft_process_logcommit_latency_bucket{store="1",le="344063"} 3.1962727e+07
raft_process_logcommit_latency_bucket{store="1",le="360447"} 3.2773494e+07
raft_process_logcommit_latency_bucket{store="1",le="376831"} 3.349722e+07
raft_process_logcommit_latency_bucket{store="1",le="393215"} 3.4152156e+07
raft_process_logcommit_latency_bucket{store="1",le="409599"} 3.4749902e+07
raft_process_logcommit_latency_bucket{store="1",le="425983"} 3.5299036e+07
raft_process_logcommit_latency_bucket{store="1",le="442367"} 3.5809502e+07
raft_process_logcommit_latency_bucket{store="1",le="458751"} 3.6288212e+07
raft_process_logcommit_latency_bucket{store="1",le="475135"} 3.6740142e+07
raft_process_logcommit_latency_bucket{store="1",le="491519"} 3.7166692e+07
raft_process_logcommit_latency_bucket{store="1",le="507903"} 3.7575906e+07
raft_process_logcommit_latency_bucket{store="1",le="524287"} 3.7968504e+07
raft_process_logcommit_latency_bucket{store="1",le="557055"} 3.8718142e+07
raft_process_logcommit_latency_bucket{store="1",le="589823"} 3.9427257e+07
raft_process_logcommit_latency_bucket{store="1",le="622591"} 4.0108505e+07
raft_process_logcommit_latency_bucket{store="1",le="655359"} 4.0765698e+07
raft_process_logcommit_latency_bucket{store="1",le="688127"} 4.1406246e+07
raft_process_logcommit_latency_bucket{store="1",le="720895"} 4.2033124e+07
raft_process_logcommit_latency_bucket{store="1",le="753663"} 4.2650642e+07
raft_process_logcommit_latency_bucket{store="1",le="786431"} 4.3261571e+07
raft_process_logcommit_latency_bucket{store="1",le="819199"} 4.3869268e+07
raft_process_logcommit_latency_bucket{store="1",le="851967"} 4.4474998e+07
raft_process_logcommit_latency_bucket{store="1",le="884735"} 4.5083851e+07
raft_process_logcommit_latency_bucket{store="1",le="917503"} 4.570077e+07
raft_process_logcommit_latency_bucket{store="1",le="950271"} 4.6326978e+07
raft_process_logcommit_latency_bucket{store="1",le="983039"} 4.6970122e+07
raft_process_logcommit_latency_bucket{store="1",le="1.015807e+06"} 4.7640341e+07
raft_process_logcommit_latency_bucket{store="1",le="1.048575e+06"} 4.8341278e+07
raft_process_logcommit_latency_bucket{store="1",le="1.114111e+06"} 4.9783408e+07
raft_process_logcommit_latency_bucket{store="1",le="1.179647e+06"} 5.1341135e+07
raft_process_logcommit_latency_bucket{store="1",le="1.245183e+06"} 5.2991061e+07
raft_process_logcommit_latency_bucket{store="1",le="1.310719e+06"} 5.4469398e+07
raft_process_logcommit_latency_bucket{store="1",le="1.376255e+06"} 5.5686873e+07
raft_process_logcommit_latency_bucket{store="1",le="1.441791e+06"} 5.6732464e+07
raft_process_logcommit_latency_bucket{store="1",le="1.507327e+06"} 5.768464e+07
raft_process_logcommit_latency_bucket{store="1",le="1.572863e+06"} 5.8582059e+07
raft_process_logcommit_latency_bucket{store="1",le="1.638399e+06"} 5.9444001e+07
raft_process_logcommit_latency_bucket{store="1",le="1.703935e+06"} 6.0283163e+07
raft_process_logcommit_latency_bucket{store="1",le="1.769471e+06"} 6.1106253e+07
raft_process_logcommit_latency_bucket{store="1",le="1.835007e+06"} 6.1920434e+07
raft_process_logcommit_latency_bucket{store="1",le="1.900543e+06"} 6.2729306e+07
raft_process_logcommit_latency_bucket{store="1",le="1.966079e+06"} 6.3534969e+07
raft_process_logcommit_latency_bucket{store="1",le="2.031615e+06"} 6.4342185e+07
raft_process_logcommit_latency_bucket{store="1",le="2.097151e+06"} 6.5147934e+07
raft_process_logcommit_latency_bucket{store="1",le="2.228223e+06"} 6.677592e+07
raft_process_logcommit_latency_bucket{store="1",le="2.359295e+06"} 6.8497692e+07
raft_process_logcommit_latency_bucket{store="1",le="2.490367e+06"} 7.0419865e+07
raft_process_logcommit_latency_bucket{store="1",le="2.621439e+06"} 7.2561709e+07
raft_process_logcommit_latency_bucket{store="1",le="2.752511e+06"} 7.4580391e+07
raft_process_logcommit_latency_bucket{store="1",le="2.883583e+06"} 7.6418251e+07
raft_process_logcommit_latency_bucket{store="1",le="3.014655e+06"} 7.8005806e+07
raft_process_logcommit_latency_bucket{store="1",le="3.145727e+06"} 7.9310908e+07
raft_process_logcommit_latency_bucket{store="1",le="3.276799e+06"} 8.0423171e+07
raft_process_logcommit_latency_bucket{store="1",le="3.407871e+06"} 8.1405717e+07
raft_process_logcommit_latency_bucket{store="1",le="3.538943e+06"} 8.2288049e+07
raft_process_logcommit_latency_bucket{store="1",le="3.670015e+06"} 8.3087228e+07
raft_process_logcommit_latency_bucket{store="1",le="3.801087e+06"} 8.3803421e+07
raft_process_logcommit_latency_bucket{store="1",le="3.932159e+06"} 8.4439888e+07
raft_process_logcommit_latency_bucket{store="1",le="4.063231e+06"} 8.5012699e+07
raft_process_logcommit_latency_bucket{store="1",le="4.194303e+06"} 8.5532891e+07
raft_process_logcommit_latency_bucket{store="1",le="4.456447e+06"} 8.6450591e+07
raft_process_logcommit_latency_bucket{store="1",le="4.718591e+06"} 8.7237357e+07
raft_process_logcommit_latency_bucket{store="1",le="4.980735e+06"} 8.7924616e+07
raft_process_logcommit_latency_bucket{store="1",le="5.242879e+06"} 8.8527797e+07
raft_process_logcommit_latency_bucket{store="1",le="5.505023e+06"} 8.9057091e+07
raft_process_logcommit_latency_bucket{store="1",le="5.767167e+06"} 8.9518254e+07
raft_process_logcommit_latency_bucket{store="1",le="6.029311e+06"} 8.9922162e+07
raft_process_logcommit_latency_bucket{store="1",le="6.291455e+06"} 9.0277154e+07
raft_process_logcommit_latency_bucket{store="1",le="6.553599e+06"} 9.0587483e+07
raft_process_logcommit_latency_bucket{store="1",le="6.815743e+06"} 9.0861642e+07
raft_process_logcommit_latency_bucket{store="1",le="7.077887e+06"} 9.1106866e+07
raft_process_logcommit_latency_bucket{store="1",le="7.340031e+06"} 9.13253e+07
raft_process_logcommit_latency_bucket{store="1",le="7.602175e+06"} 9.1520567e+07
raft_process_logcommit_latency_bucket{store="1",le="7.864319e+06"} 9.1697139e+07
raft_process_logcommit_latency_bucket{store="1",le="8.126463e+06"} 9.185804e+07
raft_process_logcommit_latency_bucket{store="1",le="8.388607e+06"} 9.200324e+07
raft_process_logcommit_latency_bucket{store="1",le="8.912895e+06"} 9.2258683e+07
raft_process_logcommit_latency_bucket{store="1",le="9.437183e+06"} 9.2473735e+07
raft_process_logcommit_latency_bucket{store="1",le="9.961471e+06"} 9.2656195e+07
raft_process_logcommit_latency_bucket{store="1",le="1.0485759e+07"} 9.28139e+07
raft_process_logcommit_latency_bucket{store="1",le="1.1010047e+07"} 9.2950869e+07
raft_process_logcommit_latency_bucket{store="1",le="1.1534335e+07"} 9.3070632e+07
raft_process_logcommit_latency_bucket{store="1",le="1.2058623e+07"} 9.3176979e+07
raft_process_logcommit_latency_bucket{store="1",le="1.2582911e+07"} 9.3270697e+07
raft_process_logcommit_latency_bucket{store="1",le="1.3107199e+07"} 9.3355405e+07
raft_process_logcommit_latency_bucket{store="1",le="1.3631487e+07"} 9.3432737e+07
raft_process_logcommit_latency_bucket{store="1",le="1.4155775e+07"} 9.3502922e+07
raft_process_logcommit_latency_bucket{store="1",le="1.4680063e+07"} 9.3566943e+07
raft_process_logcommit_latency_bucket{store="1",le="1.5204351e+07"} 9.3625122e+07
raft_process_logcommit_latency_bucket{store="1",le="1.5728639e+07"} 9.3677836e+07
raft_process_logcommit_latency_bucket{store="1",le="1.6252927e+07"} 9.3726016e+07
raft_process_logcommit_latency_bucket{store="1",le="1.6777215e+07"} 9.376919e+07
raft_process_logcommit_latency_bucket{store="1",le="1.7825791e+07"} 9.3841885e+07
raft_process_logcommit_latency_bucket{store="1",le="1.8874367e+07"} 9.3900229e+07
raft_process_logcommit_latency_bucket{store="1",le="1.9922943e+07"} 9.3946008e+07
raft_process_logcommit_latency_bucket{store="1",le="2.0971519e+07"} 9.3982452e+07
raft_process_logcommit_latency_bucket{store="1",le="2.2020095e+07"} 9.4012225e+07
raft_process_logcommit_latency_bucket{store="1",le="2.3068671e+07"} 9.4036279e+07
raft_process_logcommit_latency_bucket{store="1",le="2.4117247e+07"} 9.405677e+07
raft_process_logcommit_latency_bucket{store="1",le="2.5165823e+07"} 9.4075807e+07
raft_process_logcommit_latency_bucket{store="1",le="2.6214399e+07"} 9.4090172e+07
raft_process_logcommit_latency_bucket{store="1",le="2.7262975e+07"} 9.4101185e+07
raft_process_logcommit_latency_bucket{store="1",le="2.8311551e+07"} 9.4109665e+07
raft_process_logcommit_latency_bucket{store="1",le="2.9360127e+07"} 9.4116475e+07
raft_process_logcommit_latency_bucket{store="1",le="3.0408703e+07"} 9.412193e+07
raft_process_logcommit_latency_bucket{store="1",le="3.1457279e+07"} 9.412625e+07
raft_process_logcommit_latency_bucket{store="1",le="3.2505855e+07"} 9.4129645e+07
raft_process_logcommit_latency_bucket{store="1",le="3.3554431e+07"} 9.41326e+07
raft_process_logcommit_latency_bucket{store="1",le="3.5651583e+07"} 9.4137015e+07
raft_process_logcommit_latency_bucket{store="1",le="3.7748735e+07"} 9.4140343e+07
raft_process_logcommit_latency_bucket{store="1",le="3.9845887e+07"} 9.4142862e+07
raft_process_logcommit_latency_bucket{store="1",le="4.1943039e+07"} 9.4145016e+07
raft_process_logcommit_latency_bucket{store="1",le="4.4040191e+07"} 9.4147002e+07
raft_process_logcommit_latency_bucket{store="1",le="4.6137343e+07"} 9.4149369e+07
raft_process_logcommit_latency_bucket{store="1",le="4.8234495e+07"} 9.415633e+07
raft_process_logcommit_latency_bucket{store="1",le="5.0331647e+07"} 9.4170749e+07
raft_process_logcommit_latency_bucket{store="1",le="5.2428799e+07"} 9.4173576e+07
raft_process_logcommit_latency_bucket{store="1",le="5.4525951e+07"} 9.4174131e+07
raft_process_logcommit_latency_bucket{store="1",le="5.6623103e+07"} 9.4174477e+07
raft_process_logcommit_latency_bucket{store="1",le="5.8720255e+07"} 9.4174721e+07
raft_process_logcommit_latency_bucket{store="1",le="6.0817407e+07"} 9.4174968e+07
raft_process_logcommit_latency_bucket{store="1",le="6.2914559e+07"} 9.4175167e+07
raft_process_logcommit_latency_bucket{store="1",le="6.5011711e+07"} 9.4175344e+07
raft_process_logcommit_latency_bucket{store="1",le="6.7108863e+07"} 9.4175507e+07
raft_process_logcommit_latency_bucket{store="1",le="7.1303167e+07"} 9.4175851e+07
raft_process_logcommit_latency_bucket{store="1",le="7.5497471e+07"} 9.4176207e+07
raft_process_logcommit_latency_bucket{store="1",le="7.9691775e+07"} 9.4176257e+07
raft_process_logcommit_latency_bucket{store="1",le="8.3886079e+07"} 9.417629e+07
raft_process_logcommit_latency_bucket{store="1",le="8.8080383e+07"} 9.4176332e+07
raft_process_logcommit_latency_bucket{store="1",le="9.2274687e+07"} 9.4176391e+07
raft_process_logcommit_latency_bucket{store="1",le="9.6468991e+07"} 9.4176416e+07
raft_process_logcommit_latency_bucket{store="1",le="1.00663295e+08"} 9.4176458e+07
raft_process_logcommit_latency_bucket{store="1",le="1.04857599e+08"} 9.4176478e+07
raft_process_logcommit_latency_bucket{store="1",le="1.09051903e+08"} 9.4176491e+07
raft_process_logcommit_latency_bucket{store="1",le="1.13246207e+08"} 9.4176503e+07
raft_process_logcommit_latency_bucket{store="1",le="1.17440511e+08"} 9.4176521e+07
raft_process_logcommit_latency_bucket{store="1",le="1.21634815e+08"} 9.4176549e+07
raft_process_logcommit_latency_bucket{store="1",le="1.25829119e+08"} 9.4176572e+07
raft_process_logcommit_latency_bucket{store="1",le="1.30023423e+08"} 9.4176584e+07
raft_process_logcommit_latency_bucket{store="1",le="1.34217727e+08"} 9.4176602e+07
raft_process_logcommit_latency_bucket{store="1",le="1.42606335e+08"} 9.4176625e+07
raft_process_logcommit_latency_bucket{store="1",le="1.50994943e+08"} 9.4176647e+07
raft_process_logcommit_latency_bucket{store="1",le="1.59383551e+08"} 9.4176652e+07
raft_process_logcommit_latency_bucket{store="1",le="1.67772159e+08"} 9.4176666e+07
raft_process_logcommit_latency_bucket{store="1",le="1.76160767e+08"} 9.4176671e+07
raft_process_logcommit_latency_bucket{store="1",le="1.84549375e+08"} 9.4176672e+07
raft_process_logcommit_latency_bucket{store="1",le="1.92937983e+08"} 9.417668e+07
raft_process_logcommit_latency_bucket{store="1",le="2.01326591e+08"} 9.4176681e+07
raft_process_logcommit_latency_bucket{store="1",le="+Inf"} 9.4176681e+07
raft_process_logcommit_latency_sum{store="1"} 1.71643585649239e+14
raft_process_logcommit_latency_count{store="1"} 9.4176681e+07
`
	output := `# HELP raft_process_logcommit_latency Latency histogram for committing Raft log entries
# TYPE raft_process_logcommit_latency histogram
raft_process_logcommit_latency_bucket{store="1",le="70000"} 8
raft_process_logcommit_latency_bucket{store="1",le="80000"} 2513
raft_process_logcommit_latency_bucket{store="1",le="90000"} 33293
raft_process_logcommit_latency_bucket{store="1",le="100000"} 194901
raft_process_logcommit_latency_bucket{store="1",le="200000"} 1.6427827e+07
raft_process_logcommit_latency_bucket{store="1",le="300000"} 2.9168318e+07
raft_process_logcommit_latency_bucket{store="1",le="400000"} 3.4399697e+07
raft_process_logcommit_latency_bucket{store="1",le="500000"} 3.7378518e+07
raft_process_logcommit_latency_bucket{store="1",le="600000"} 3.9638838e+07
raft_process_logcommit_latency_bucket{store="1",le="700000"} 4.1633386e+07
raft_process_logcommit_latency_bucket{store="1",le="800000"} 4.3513215e+07
raft_process_logcommit_latency_bucket{store="1",le="900000"} 4.5371244e+07
raft_process_logcommit_latency_bucket{store="1",le="1e+06"} 4.7317034e+07
raft_process_logcommit_latency_bucket{store="1",le="2e+06"} 6.3952779e+07
raft_process_logcommit_latency_bucket{store="1",le="3e+06"} 7.7828304e+07
raft_process_logcommit_latency_bucket{store="1",le="4e+06"} 8.4736367e+07
raft_process_logcommit_latency_bucket{store="1",le="5e+06"} 8.7968944e+07
raft_process_logcommit_latency_bucket{store="1",le="6e+06"} 8.9877e+07
raft_process_logcommit_latency_bucket{store="1",le="7e+06"} 9.1034007e+07
raft_process_logcommit_latency_bucket{store="1",le="8e+06"} 9.1780419e+07
raft_process_logcommit_latency_bucket{store="1",le="9e+06"} 9.2294412e+07
raft_process_logcommit_latency_bucket{store="1",le="1e+07"} 9.2667785e+07
raft_process_logcommit_latency_bucket{store="1",le="2e+07"} 9.3948687e+07
raft_process_logcommit_latency_bucket{store="1",le="3e+07"} 9.4119804e+07
raft_process_logcommit_latency_bucket{store="1",le="4e+07"} 9.4143021e+07
raft_process_logcommit_latency_bucket{store="1",le="5e+07"} 9.4168469e+07
raft_process_logcommit_latency_bucket{store="1",le="6e+07"} 9.4174872e+07
raft_process_logcommit_latency_bucket{store="1",le="7e+07"} 9.4175745e+07
raft_process_logcommit_latency_bucket{store="1",le="8e+07"} 9.417626e+07
raft_process_logcommit_latency_bucket{store="1",le="9e+07"} 9.417636e+07
raft_process_logcommit_latency_bucket{store="1",le="1e+08"} 9.4176452e+07
raft_process_logcommit_latency_bucket{store="1",le="2e+08"} 9.4176681e+07
raft_process_logcommit_latency_bucket{store="1",le="3e+08"} 9.4176681e+07
raft_process_logcommit_latency_bucket{store="1",le="+Inf"} 9.4176681e+07
raft_process_logcommit_latency_sum{store="1"} 1.71643585649239e+14
raft_process_logcommit_latency_count{store="1"} 9.4176681e+07
`
	var parser expfmt.TextParser
	config := &BucketConfig{
		Startns: 100,
		Bins:    10}

	metricFamilies, _ := parser.TextToMetricFamilies(strings.NewReader(input))
	for _, mf := range metricFamilies {
		TranslateHistogram(config, mf)
		var buf bytes.Buffer
		expfmt.MetricFamilyToText(&buf, mf)
		res := (buf.String() == output)
		//fmt.Println(buf.String())
		if !res {
			t.Fatal("TranslateHistogram failed")
		}

	}

}

func TestIdentityConversion(t *testing.T) {

	input := `# HELP raft_process_logcommit_latency Latency histogram for committing Raft log entries
# TYPE raft_process_logcommit_latency histogram
raft_process_logcommit_latency_bucket{store="1",le="70000"} 8
raft_process_logcommit_latency_bucket{store="1",le="80000"} 2513
raft_process_logcommit_latency_bucket{store="1",le="90000"} 33293
raft_process_logcommit_latency_bucket{store="1",le="100000"} 194901
raft_process_logcommit_latency_bucket{store="1",le="200000"} 1.6427827e+07
raft_process_logcommit_latency_bucket{store="1",le="300000"} 2.9168318e+07
raft_process_logcommit_latency_bucket{store="1",le="400000"} 3.4399697e+07
raft_process_logcommit_latency_bucket{store="1",le="500000"} 3.7378518e+07
raft_process_logcommit_latency_bucket{store="1",le="600000"} 3.9638838e+07
raft_process_logcommit_latency_bucket{store="1",le="700000"} 4.1633386e+07
raft_process_logcommit_latency_bucket{store="1",le="800000"} 4.3513215e+07
raft_process_logcommit_latency_bucket{store="1",le="900000"} 4.5371244e+07
raft_process_logcommit_latency_bucket{store="1",le="1e+06"} 4.7317034e+07
raft_process_logcommit_latency_bucket{store="1",le="2e+06"} 6.3952779e+07
raft_process_logcommit_latency_bucket{store="1",le="3e+06"} 7.7828304e+07
raft_process_logcommit_latency_bucket{store="1",le="4e+06"} 8.4736367e+07
raft_process_logcommit_latency_bucket{store="1",le="5e+06"} 8.7968944e+07
raft_process_logcommit_latency_bucket{store="1",le="6e+06"} 8.9877e+07
raft_process_logcommit_latency_bucket{store="1",le="7e+06"} 9.1034007e+07
raft_process_logcommit_latency_bucket{store="1",le="8e+06"} 9.1780419e+07
raft_process_logcommit_latency_bucket{store="1",le="9e+06"} 9.2294412e+07
raft_process_logcommit_latency_bucket{store="1",le="1e+07"} 9.2667785e+07
raft_process_logcommit_latency_bucket{store="1",le="2e+07"} 9.3948687e+07
raft_process_logcommit_latency_bucket{store="1",le="3e+07"} 9.4119804e+07
raft_process_logcommit_latency_bucket{store="1",le="4e+07"} 9.4143021e+07
raft_process_logcommit_latency_bucket{store="1",le="5e+07"} 9.4168469e+07
raft_process_logcommit_latency_bucket{store="1",le="6e+07"} 9.4174872e+07
raft_process_logcommit_latency_bucket{store="1",le="7e+07"} 9.4175745e+07
raft_process_logcommit_latency_bucket{store="1",le="8e+07"} 9.417626e+07
raft_process_logcommit_latency_bucket{store="1",le="9e+07"} 9.417636e+07
raft_process_logcommit_latency_bucket{store="1",le="1e+08"} 9.4176452e+07
raft_process_logcommit_latency_bucket{store="1",le="2e+08"} 9.4176681e+07
raft_process_logcommit_latency_bucket{store="1",le="3e+08"} 9.4176681e+07
raft_process_logcommit_latency_bucket{store="1",le="+Inf"} 9.4176681e+07
raft_process_logcommit_latency_sum{store="1"} 1.71643585649239e+14
raft_process_logcommit_latency_count{store="1"} 9.4176681e+07
`

	var parser expfmt.TextParser
	config := &BucketConfig{
		Startns: 100,
		Bins:    10}

	metricFamilies, _ := parser.TextToMetricFamilies(strings.NewReader(input))
	for _, mf := range metricFamilies {
		TranslateHistogram(config, mf)
		var buf bytes.Buffer
		expfmt.MetricFamilyToText(&buf, mf)
		res := (buf.String() == input)

		if !res {
			t.Fatal("TranslateHistogram failed")
		}

	}

}
