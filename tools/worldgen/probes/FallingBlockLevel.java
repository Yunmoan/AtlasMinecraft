import java.util.*;

/** Defined level for actual FallingBlock placement, shapes and scheduled ticks. */
public class FallingBlockLevel extends aqu {
    public Map<jd,dtc> blocks;
    public List<String> events;
    public List<bsr> spawned;
    public ayw random;
    public dwy dimension;
    public jd position;
    public dcs rules;
    public dut border;
    public ka access;
    public brl damage;
    public aqs chunkSource;
    public dql brushEntity;
    public long brushTime;
    public boolean brushProbe;
    public boolean kelpProbe, kelpDestroyed;
    public FallingBlockLevel(){super(null,null,null,null,null,null,null,false,0L,null,false,null);}
    @Override public int I_(){return -64;}
    @Override public int am(){return 320;}
    @Override public dwy D_(){return dimension;}
    @Override public cpl J(){return cpn.g;}
    @Override public ayw E_(){return random;}
    @Override public ka H_(){return access;}
    @Override public brl aj(){return damage;}
    @Override public aqs l(){return chunkSource;}
    @Override public bnf ag(){return bnc.a;}
    @Override public dcs ab(){return rules;}
    @Override public dut C_(){return border;}
    @Override public dcc c(int x,int z){return this;}
    @Override public boolean b(int x,int z){return true;}
    @Override public boolean p(jd p){return true;}
    @Override public long Z(){return brushTime;}
    @Override public dqh c_(jd p){return p.equals(position)&&a_(p).b() instanceof dgd?brushEntity:null;}
    @Override public boolean a(jd p,boolean drops,bsr source,int limit){
        if(!kelpProbe)return super.a(p,drops,source,limit);
        if(!p.equals(position)||!drops||source!=null||limit!=512)throw new IllegalStateException("Unexpected kelp destruction boundary");
        kelpDestroyed=true;events.add("destroy "+p.u()+" "+p.v()+" "+p.w()+" true 512");return true;
    }
    @Override public void c(int event,jd p,int data){events.add("level_event "+event+" "+p.u()+" "+p.v()+" "+p.w()+" "+data);}
    @Override public void a(jm<dxz> event,exc p,dxz.a context){
        String key=event.e().orElseThrow().a().toString();
        events.add("game_event "+key+(brushProbe&&key.equals("minecraft:block_destroy")?String.format(Locale.ROOT," %.9f %.9f %.9f",p.c,p.d,p.e):""));
    }
    @Override public List<exv> c(bsr entity,ewx box){return List.of();}
    @Override public Iterable<exv> e(bsr entity,ewx box){
        var shapes=new ArrayList<exv>();
        for(var entry:blocks.entrySet()){jd p=entry.getKey();exv s=entry.getValue().k(this,p);if(!s.c())shapes.add(s.a(p.u(),p.v(),p.w()));}
        return shapes;
    }
    @Override public dtc a_(jd p){return blocks.getOrDefault(p,dga.a.o());}
    @Override public epe b_(jd p){return a_(p).u();}
    @Override public boolean a(jd p,dtc s,int flags){
        if(a_(p)==s)return false;
        blocks.put(p.j(),s);events.add("set "+p.u()+" "+p.v()+" "+p.w()+" "+dfy.i(s)+" "+flags);
        if(brushEntity!=null&&p.equals(position)&&s.b() instanceof dgd)brushEntity.c(s);
        return true;
    }
    @Override public void a(jd p,dfy b,int delay){events.add("block_tick "+p.u()+" "+p.v()+" "+p.w()+" "+lt.e.b(b)+" "+delay+" 0");}
    @Override public boolean b(bsr entity){spawned.add(entity);events.add("spawn "+lt.f.b(entity.am())+" "+entity.dt()+" "+entity.dv()+" "+entity.dz());return true;}
    public static int[] states(){
        var ids=new ArrayList<Integer>();for(int id=0;id<26684;id++)if(dfy.a(id).b() instanceof diu)ids.add(id);
        return ids.stream().mapToInt(Integer::intValue).toArray();
    }
    public void configure(int id,int below,int height,long seed){
        blocks=new LinkedHashMap<>();events=new ArrayList<>();spawned=new ArrayList<>();random=new dyz(seed);
        brushEntity=null;brushTime=0;
        position=new jd(-17,height,47);blocks.put(position,dfy.a(id));blocks.put(position.e(),dfy.a(below));
    }
    public static int[] brushStates(){
        var ids=new ArrayList<Integer>();for(int id=0;id<26684;id++)if(dfy.a(id).b() instanceof dgd)ids.add(id);
        return ids.stream().mapToInt(Integer::intValue).toArray();
    }
    public static int[] kelpStates(){
        var ids=new ArrayList<Integer>();for(int id=0;id<26684;id++)if(dfy.a(id).b() instanceof dkc||dfy.a(id).b() instanceof dkd)ids.add(id);
        return ids.stream().mapToInt(Integer::intValue).toArray();
    }
    public List<String> runKelp(int id,int below,long seed,long entitySeed,boolean drops)throws Exception {
        configure(id,below,64,seed);kelpProbe=true;kelpDestroyed=false;rules=new dcs();rules.a(dcs.h).a(drops,(net.minecraft.server.MinecraftServer)null);
        var field=dcw.class.getDeclaredField("z");field.setAccessible(true);field.set(this,random);
        System.setProperty("atlas.entitySeed",Long.toString(entitySeed));
        a_(position).a(this,position,random); // original GrowingPlantBlock.tick
        if(kelpDestroyed)dfy.a(this,position,new cuq(lt.g.a(akr.a("minecraft:kelp")))); // original Block.popResource
        return fullEntityTags();
    }
    public String kelpItemNext()throws Exception {
        if(spawned.isEmpty())return "";
        var field=bsr.class.getDeclaredField("ah");field.setAccessible(true);
        return Long.toUnsignedString(((ayw)field.get(spawned.get(0))).g());
    }
    public void configureBrush(int id,int below,int height,int count,long now)throws Exception {
        configure(id,below,height,1234L);brushProbe=true;brushTime=now;
        if(count<0)return;
        brushEntity=new dql(position,dfy.a(id));brushEntity.a(this);
        for(var entry:Map.of("i",(Object)count,"j",100L,"k",150L,"m",ji.c).entrySet()){
            var field=dql.class.getDeclaredField(entry.getKey());field.setAccessible(true);field.set(brushEntity,entry.getValue());
        }
    }
    public List<Object> brushProgress()throws Exception {
        var result=new ArrayList<Object>();if(brushEntity==null)return result;
        for(String key:new String[]{"i","j","k","m"}){
            var field=dql.class.getDeclaredField(key);field.setAccessible(true);Object value=field.get(brushEntity);
            result.add(key.equals("m")?(value==null?-1:((ji)value).ordinal()):value);
        }
        return result;
    }
    public long run(String action,int face){
        dtc s=a_(position);
        if(action.equals("tick"))s.a(this,position,random);
        else if(action.equals("place"))s.a(this,position,dga.a.o(),false);
        else {ji d=ji.values()[face];dtc after=s.a(d,a_(position.a(d)),this,position,position.a(d));events.add("result "+dfy.i(after));}
        return random.g();
    }
    public List<String> entityTags()throws Exception{
        var result=new ArrayList<String>();
        for(bsr e:spawned){ub n=new ub();e.f(n);n.r("UUID");var bytes=new java.io.ByteArrayOutputStream();uo.a((uy)n,(java.io.DataOutput)new java.io.DataOutputStream(bytes));result.add(HexFormat.of().formatHex(bytes.toByteArray()));}
        return result;
    }
    public void configureMotion(int id,int support){
        configure(id,0,64,1234L);rules=new dcs();border=new dut();
        for(int x=-19;x<=-15;x++)for(int z=45;z<=49;z++)blocks.put(new jd(x,60,z),dfy.a(support));
        a_(position).a(this,position,random);events.clear();
    }
    public List<String> motionTick()throws Exception{
        for(bsr e:new ArrayList<>(spawned))if(!e.dJ())e.l();
        return entityTags();
    }
    public List<Boolean> removed(){return spawned.stream().map(bsr::dJ).toList();}
    public void configureDrop(int id,int support,long seed,long entitySeed,boolean drops,boolean dropItem,boolean cancel,int age,double y)throws Exception {
        System.setProperty("atlas.entitySeed",Long.toString(entitySeed^0x7465727261696eL));
        configureMotion(id,support);
        random=new dyz(seed);java.lang.reflect.Field field=dcw.class.getDeclaredField("z");field.setAccessible(true);field.set(this,random);
        rules.a(dcs.i).a(drops,(net.minecraft.server.MinecraftServer)null);
        cjg e=(cjg)spawned.get(0);ub tag=new ub();e.f(tag);
        tag.a("Time",age);tag.a("DropItem",dropItem);tag.a("CancelDrop",cancel);
        e.g(tag);e.a_(-16.5,y,47.5);
        System.setProperty("atlas.entitySeed",Long.toString(entitySeed));
        events.clear();
    }
    public List<String> dropTick()throws Exception {
        spawned.get(0).l();
        return fullEntityTags();
    }
    public List<String> fullEntityTags()throws Exception {
        var result=new ArrayList<String>();
        for(bsr e:spawned){ub n=new ub();e.f(n);n.a("id",lt.f.b(e.am()).toString());var bytes=new java.io.ByteArrayOutputStream();uo.a((uy)n,(java.io.DataOutput)new java.io.DataOutputStream(bytes));result.add(HexFormat.of().formatHex(bytes.toByteArray()));}
        return result;
    }
    public String itemRandomNext()throws Exception {
        if(spawned.size()==1)return "";
        java.lang.reflect.Field field=bsr.class.getDeclaredField("ah");field.setAccessible(true);
        return Long.toUnsignedString(((ayw)field.get(spawned.get(1))).g());
    }
    public static int namedState(String name){return dfy.i(lt.e.a(akr.a("minecraft:"+name)).o());}
}
