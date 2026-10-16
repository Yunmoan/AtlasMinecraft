package atlasprobe;

import java.util.ArrayList;
import java.util.List;

final class OceanRuinProbeSupport {
    static String[] names() {
        List<String> names=new ArrayList<>();
        for(String material:new String[]{"warm","brick","cracked","mossy"})for(int index=1;index<=8;index++)names.add("underwater_ruin/"+material+"_"+index);
        for(int index=4;index<=7;index++)names.add("underwater_ruin/big_warm_"+index);
        for(String material:new String[]{"brick","cracked","mossy"})for(int index:new int[]{1,2,3,8})names.add("underwater_ruin/big_"+material+"_"+index);
        return names.toArray(String[]::new);
    }
}
