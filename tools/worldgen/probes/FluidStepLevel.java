import java.util.*;
import io.netty.buffer.Unpooled;

/** Defined block world for one original FluidState.tick call. Level callbacks
 * are recorded separately; this oracle does not pretend to run a server tick. */
public class FluidStepLevel extends dcw {
    public Map<jd,dtc> edits;
    public List<String> events;
    public int floor;
    public dwy dimension;
    public dcs rules;
    public ayw random;
    public int shapeBrightness = 15;
    public List<int[]> shapeInputs;
    public FluidStepLevel() { super(null,null,null,null,null,false,false,0L,0); }
    @Override public void a(jd p,dtc before,dtc after,int flags) {throw new UnsupportedOperationException();}
    @Override public void a(cmx player,double x,double y,double z,jm<avo> sound,avq category,float volume,float pitch,long seed) {events.add("sound "+lt.b.b(sound.a())+" "+category.ordinal()+" "+seed);}
    @Override public void a(cmx player,bsr entity,jm<avo> sound,avq category,float volume,float pitch,long seed) {throw new UnsupportedOperationException();}
    @Override public String I() {return "FluidStepLevel";}
    @Override public bsr a(int id) {throw new UnsupportedOperationException();}
    @Override public brc s() {throw new UnsupportedOperationException();}
    @Override public eqt a(eqr id) {throw new UnsupportedOperationException();}
    @Override public void a(eqr id,eqt data) {throw new UnsupportedOperationException();}
    @Override public eqr v() {throw new UnsupportedOperationException();}
    @Override public void a(int id,jd p,int stage) {throw new UnsupportedOperationException();}
    @Override public eyg M() {throw new UnsupportedOperationException();}
    @Override public czd r() {throw new UnsupportedOperationException();}
    @Override protected dxp<bsr> G() {throw new UnsupportedOperationException();}
    @Override public cwt K() {throw new UnsupportedOperationException();}
    @Override public java.util.List<? extends cmx> x() {return java.util.List.of();}
    @Override public cpl J() {return cpn.g;}
    @Override public jm<ddw> a(int x,int y,int z) {throw new UnsupportedOperationException();}
    @Override public float a(ji direction,boolean shade) {throw new UnsupportedOperationException();}
    @Override public dvc N() {throw new UnsupportedOperationException();}
    @Override public eyp<dfy> P() {return eym.b();}
    @Override public eyp<epd> O() {throw new UnsupportedOperationException();}
    @Override public bqp d_(jd p) {throw new UnsupportedOperationException();}
    @Override public net.minecraft.server.MinecraftServer o() {return null;}
    @Override public void a(cmx player,int event,jd p,int data) {c(event,p,data);}
    @Override public void a(jm<dxz> event,exc p,dxz.a context) {events.add("game_event "+event.e().orElseThrow().a());}
    @Override public int I_() { return -64; }
    @Override public int am() { return 320; }
    @Override public int J_() { return 384; }
    @Override public dwy D_() { return dimension; }
    @Override public dcs ab() { return rules; }
    @Override public ka H_() { return new ka.c(List.of(lt.e)); }
    @Override public ayw E_() { return random; }
    @Override public dtc a_(jd p) { return edits.getOrDefault(p,p.v()<=floor?dga.b.o():dga.a.o()); }
    @Override public epe b_(jd p) { return a_(p).u(); }
    @Override public dqh c_(jd p) { return null; }
    @Override public int b(jd p,int darken) {return shapeBrightness;}
    @Override public void a(jd p,dfy block,int delay) {events.add("block_tick "+p.u()+" "+p.v()+" "+p.w()+" "+lt.e.b(block)+" "+delay+" 0");}
    @Override public void a(jd p,dfy block,int delay,eyx priority) {events.add("block_tick "+p.u()+" "+p.v()+" "+p.w()+" "+lt.e.b(block)+" "+delay+" "+priority.a());}
    @Override public boolean a(jd p,dtc state,int flags) {
        if(a_(p)==state)return false;
        edits.put(p.j(),state);
        events.add("set "+p.u()+" "+p.v()+" "+p.w()+" "+dfy.i(state)+" "+flags);
        return true;
    }
    @Override public boolean b(jd p,dtc state) { return a(p,state,3); }
    @Override public void a(jd p,epd fluid,int delay) {
        events.add("schedule "+p.u()+" "+p.v()+" "+p.w()+" "+lt.c.b(fluid)+" "+delay);
    }
    @Override public void a(jd p,dfy block) {
        events.add("notify "+p.u()+" "+p.v()+" "+p.w()+" "+dfy.i(block.o()));
    }
    @Override public void c(int event,jd p,int data) {
        events.add("event "+p.u()+" "+p.v()+" "+p.w()+" "+event+" "+data);
    }
    @SuppressWarnings({"unchecked","rawtypes"})
    public static dtc state(String name,String...properties) {
        dfy block=lt.e.a(akr.a("minecraft:"+name));dtc state=block.o();
        for(String text:properties) {
            String[] pair=text.split("=");duf property=block.l().a(pair[0]);
            state=(dtc)state.a(property,(Comparable)property.b(pair[1]).orElseThrow());
        }
        return state;
    }
    public void configure(long seed,boolean convert,int scenario,String liquid,int level,int ox,int oy,int oz) {
        edits=new LinkedHashMap<>();events=new ArrayList<>();random=new dyz(seed);rules=new dcs();
        rules.a(dcs.h).a(false,(net.minecraft.server.MinecraftServer)null);rules.a(dcs.U).a(convert,(net.minecraft.server.MinecraftServer)null);rules.a(dcs.V).a(convert,(net.minecraft.server.MinecraftServer)null);
        floor=oy-1;dtc source=state(liquid,"level=0");
        edits.put(new jd(ox,oy,oz),state(liquid,"level="+level));
        switch(scenario) {
            case 1:floor=oy-3;break;
            case 2:floor=oy-2;put(ox,oy,oz,0,-1,0,source);break;
            case 3:put(ox,oy,oz,2,-1,0,dga.a.o());break;
            case 4:put(ox,oy,oz,2,-1,0,dga.a.o());put(ox,oy,oz,-2,-1,0,dga.a.o());break;
            case 5:put(ox,oy,oz,3,-1,0,dga.a.o());break;
            case 6:put(ox,oy,oz,1,0,0,state("oak_stairs","facing=west","half=bottom","waterlogged=false"));break;
            case 7:put(ox,oy,oz,1,0,0,source);put(ox,oy,oz,-1,0,0,source);break;
            case 8:floor=oy-3;put(ox,oy,oz,1,0,0,source);put(ox,oy,oz,-1,0,0,source);break;
            case 9:put(ox,oy,oz,0,1,0,source);break;
            case 10:put(ox,oy,oz,0,1,0,state(liquid,"level=5"));break;
            case 11:put(ox,oy,oz,1,0,0,state(liquid,"level=1"));break;
            case 12:put(ox,oy,oz,-1,0,0,source);put(ox,oy,oz,1,0,0,state(liquid,"level=4"));break;
            case 13:floor=oy-2;put(ox,oy,oz,0,-1,0,state(liquid.equals("water")?"lava":"water","level=0"));break;
            case 14:put(ox,oy,oz,-1,0,0,state("oak_stairs","facing=east","half=bottom","waterlogged=true"));break;
            case 15:put(ox,oy,oz,1,0,0,state("short_grass"));break;
            case 16:put(ox,oy,oz,1,0,0,state("oak_door"));put(ox,oy,oz,0,0,1,state("oak_sign"));break;
            case 17:put(ox,oy,oz,1,0,0,state("oak_slab","type=top","waterlogged=false"));put(ox,oy,oz,0,0,1,state("oak_slab","type=bottom","waterlogged=false"));break;
            case 18:floor=oy-3;put(ox,oy,oz,1,0,0,source);put(ox,oy,oz,-1,0,0,source);put(ox,oy,oz,0,0,1,source);break;
            case 19:put(ox,oy,oz,0,0,0,state("oak_stairs","facing=east","half=bottom","waterlogged=true"));break;
            case 20:put(ox,oy,oz,1,0,0,state("oak_slab","type=bottom","waterlogged=false"));put(ox,oy,oz,2,0,0,source);break;
            case 21:floor=oy-3;put(ox,oy,oz,1,0,0,source);put(ox,oy,oz,-1,0,0,source);put(ox,oy,oz,0,-1,0,source);break;
            case 22:put(ox,oy,oz,1,0,0,state("pointed_dripstone","thickness=base","vertical_direction=up","waterlogged=false"));put(ox,oy,oz,2,0,0,source);break;
            case 23:floor=oy-2;put(ox,oy,oz,0,-1,0,state("oak_slab","type=bottom","waterlogged=true"));break;
            case 24:put(ox,oy,oz,0,1,0,source);put(ox,oy,oz,1,0,0,state(liquid,"level=1"));break;
            case 25:put(ox,oy,oz,0,0,0,state("pointed_dripstone","thickness=base","vertical_direction=up","waterlogged=true"));put(ox,oy,oz,-1,0,0,state("sticky_piston","extended=true","facing=north"));break;
            case 26:put(ox,oy,oz,1,0,0,state(liquid.equals("water")?"lava":"water","level=4"));break;
            case 27:put(ox,oy,oz,1,0,0,state(liquid.equals("water")?"lava":"water","level=5"));break;
        }
    }
    void put(int x,int y,int z,int dx,int dy,int dz,dtc state) { edits.put(new jd(x+dx,y+dy,z+dz),state); }
    public List<int[]> inputs() {
        List<int[]> result=new ArrayList<>();for(var e:edits.entrySet())result.add(new int[]{e.getKey().u(),e.getKey().v(),e.getKey().w(),dfy.i(e.getValue())});return result;
    }
    public long step(int x,int y,int z) {jd p=new jd(x,y,z);a_(p).u().a(this,p);return random.g();}
    public void configureLiquidHook(int scene,String liquid,int level,int x,int y,int z) {
        configure(0,true,0,liquid,level,x,y,z);floor=y-2;
        ji[] directions=ji.values();
        if(scene>=1 && scene<=6)edits.put(new jd(x,y,z).a(directions[scene-1]),state("water","level=3"));
        if(scene>=7 && scene<=12)edits.put(new jd(x,y,z).a(directions[scene-7]),state("oak_slab","waterlogged=true","type=bottom"));
        if(scene>=13 && scene<=18){edits.put(new jd(x,y-1,z),state("soul_soil"));edits.put(new jd(x,y,z).a(directions[scene-13]),state("blue_ice"));}
        if(scene==19){edits.put(new jd(x,y-1,z),state("soul_soil"));edits.put(new jd(x,y,z-1),state("blue_ice"));edits.put(new jd(x,y,z+1),state("water"));}
        if(scene==20){edits.put(new jd(x,y-1,z),state("soul_soil"));edits.put(new jd(x,y,z-1),state("blue_ice"));edits.put(new jd(x,y+1,z),state("water"));}
        if(scene==21){edits.put(new jd(x,y-1,z),state("water"));edits.put(new jd(x,y,z-1),state("blue_ice"));}
        if(scene==22){edits.put(new jd(x,y-1,z),state("soul_sand"));edits.put(new jd(x,y,z-1),state("blue_ice"));}
        if(scene==23)edits.put(new jd(x+1,y,z),state("lava"));
        if(scene==24)edits.put(new jd(x+1,y,z),state("water"));
        if(scene==25)edits.put(new jd(x+1,y,z),state("water","level=9"));
    }
    public void liquidHook(int hook,int face,int x,int y,int z)throws Exception {
        jd p=new jd(x,y,z);dtc state=a_(p);Object block=state.b();java.lang.reflect.Method method;
        if(hook==0){method=dko.class.getDeclaredMethod("b",dtc.class,dcw.class,jd.class,dtc.class,boolean.class);method.setAccessible(true);method.invoke(block,state,this,p,dga.a.o(),false);}
        else if(hook==1){method=dko.class.getDeclaredMethod("a",dtc.class,dcw.class,jd.class,dfy.class,jd.class,boolean.class);method.setAccessible(true);method.invoke(block,state,this,p,dga.b,p.a(ji.values()[face]),false);}
        else {jd q=p.a(ji.values()[face]);method=dko.class.getDeclaredMethod("a",dtc.class,ji.class,dtc.class,dcx.class,jd.class,jd.class);method.setAccessible(true);dtc result=(dtc)method.invoke(block,state,ji.values()[face],a_(q),this,p,q);if(result!=state)throw new IllegalStateException("LiquidBlock.updateShape changed state");}
    }
    public void configureContainer(int id,long seed)throws Exception {
        edits=new LinkedHashMap<>();events=new ArrayList<>();floor=63;random=new dyz(seed);rules=new dcs();
        edits.put(new jd(8,64,8),dfy.a(id));
        java.lang.reflect.Field f=dcw.class.getDeclaredField("f");f.setAccessible(true);f.set(this,random);
    }
    public boolean placeContainer(int kind) {
        epd[] fluids={epf.a,epf.b,epf.c,epf.d,epf.e};jd p=new jd(8,64,8);dtc state=a_(p);
        return ((dkp)state.b()).a(this,p,state,fluids[kind].g());
    }
    public static Map<String,Object> containerSounds() {
        if(lt.j.a(dqj.G)!=32)throw new IllegalStateException("Campfire block entity registry ID changed");
        Map<String,Object> result=new TreeMap<>();
        for(avo sound:new avo[]{avp.kf,avp.dL})result.put(lt.b.b(sound).toString(),new Object[]{lt.b.a(sound),sound.a(1.0f)});
        return result;
    }
    public static List<String> containerSoundPackets()throws Exception {
        List<String> result=new ArrayList<>();
        for(avo sound:new avo[]{avp.kf,avp.dL}) {
            dyz rng=new dyz(1234L);afp packet=new afp(lt.b.e(sound),avq.e,-38.5,-52.5,712.5,1.0f,1.0f,rng.g());
            var bytes=Unpooled.buffer();wk buffer=new wk(bytes,new ka.c(List.of(lt.b)));
            java.lang.reflect.Method write=afp.class.getDeclaredMethod("a",wk.class);write.setAccessible(true);write.invoke(packet,buffer);
            byte[] encoded=new byte[bytes.readableBytes()];bytes.readBytes(encoded);bytes.release();
            result.add(lt.b.b(sound)+" "+HexFormat.of().formatHex(encoded)+" "+Long.toUnsignedString(rng.g()));
        }
        return result;
    }
    public int updateShapeCase(int id,int face,int scene)throws Exception {
        configureContainer(id,id*31L+scene*6L+face);
        shapeBrightness=15;
        jd p=new jd(8,64,8);dtc state=a_(p);ji direction=ji.values()[face];jd q=p.a(direction);
        dtc[] neighbors={dga.a.o(),dga.b.o(),state("water"),state("grass_block"),state("farmland"),state};
        edits.put(q,neighbors[scene]);
        return dfy.i(state.a(direction,neighbors[scene],this,p,q));
    }
    public static boolean plantShapeState(int id) {
        dfy b=dfy.a(id).b();return b instanceof dgh || b instanceof dfb || b instanceof djp || b instanceof dnp;
    }
    public int updatePlantShapeCase(int id,int face,int scene)throws Exception {
        configureContainer(id,id*31L+scene*6L+face);shapeBrightness=15;
        jd p=new jd(8,64,8);dtc s=a_(p);
        String[] soil={"grass_block","farmland","farmland","farmland","mycelium","stone","stone","dirt","soul_sand","soul_soil","clay","magma_block","ice","water","water","bamboo","netherrack","warped_nylium","crimson_nylium","moss_block","sand","coarse_dirt","podzol","mud","gravel"};
        edits.put(p.e(),scene==14?state("water","level=3"):state(soil[scene]));
        shapeBrightness=switch(scene){case 2->7;case 3->8;case 5->12;case 6->13;case 7->0;default->15;};
        if(scene==0 || scene==2 || scene==3 || scene==7) {
            if(s.b(dts.ae)){boolean lower=s.c(dts.ae)==dty.b;edits.put(lower?p.d():p.e(),(dtc)s.a(dts.ae,lower?dty.a:dty.b));}
            if(s.b() instanceof djp || s.b() instanceof dnp)edits.put(p.d(),dga.b.o());
            if(s.b() instanceof dkt && s.c(dkt.d))edits.put(p.d(),state("mangrove_leaves"));
        }
        if(scene==15)edits.put(p.d(),state("bamboo"));
        ji direction=ji.values()[face];jd q=p.a(direction);
        shapeInputs=inputs();
        return dfy.i(s.a(direction,a_(q),this,p,q));
    }
}
