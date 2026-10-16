package atlasprobe;

import java.util.ArrayList;
import java.util.List;

final class ShipwreckProbeSupport {
    static String[] names() {
        List<String> names=new ArrayList<>();
        for(String name:new String[]{"with_mast","upsidedown_full","upsidedown_fronthalf","upsidedown_backhalf","sideways_full","sideways_fronthalf","sideways_backhalf","rightsideup_full","rightsideup_fronthalf","rightsideup_backhalf"})names.add("shipwreck/"+name);
        for(int index=0;index<10;index++)names.add(names.get(index)+"_degraded");
        return names.toArray(String[]::new);
    }
}
