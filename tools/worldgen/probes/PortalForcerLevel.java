import java.util.*;

/** Constructor-free test level; the actual PortalForcer still executes. */
public class PortalForcerLevel extends aqu {
    public Map<jd,dtc> edits;
    public List<String> writes;
    public int minY,maxY,logical,mode;
    public dut border;
    public List<exv> collisions;
    public PortalForcerLevel() { super(null,null,null,null,null,null,null,false,0,null,false,null); }
    @Override public int I_() { return minY; }
    @Override public int am() { return maxY; }
    @Override public int k() { return logical; }
    @Override public dut C_() { return border; }
    @Override public dtc a_(jd p) {
        dtc found=edits.get(p);
        if(found!=null)return found;
        if(p.v()<minY||p.v()>=maxY)return dga.a.o();
        if(mode==2)return dga.b.o();
        if(mode==3)return dga.H.o();
        if(mode==4)return dga.a.o();
        if(mode==6&&p.v()>40&&p.v()<43)return dga.a.o();
        if(mode==6)return dga.b.o();
        if(mode==8&&p.v()>=120)return dga.b.o();
        if(p.v()<=40) {
            if(mode==1)return dga.G.o();
            if(mode==5&&p.w()!=0)return dga.a.o();
            return dga.b.o();
        }
        return dga.a.o();
    }
    @Override public int a(dyy.a map,int x,int z) {
        for(int y=maxY-1;y>=minY;y--)if(map.e().test(a_(new jd(x,y,z))))return y+1;
        return minY;
    }
    @Override public boolean a(jd p,dtc state,int flags) {
        edits.put(p.j(),state);
        writes.add(p.u()+" "+p.v()+" "+p.w()+" "+dfy.i(state));
        return true;
    }
    @Override public boolean b(jd p,dtc state) { return a(p,state,3); }
    @Override public Iterable<exv> e(bsr entity,ewx box) { return collisions; }
}
