package atlasprobe;
import java.io.*;
import java.lang.reflect.*;
public final class PortalProcessorProbe {
    public static void main(String[] args) throws Exception {
        Object version=Class.forName("t").getMethod("a").invoke(null);
        Class.forName("ab").getMethod("a",Class.forName("ae")).invoke(null,version);
        Class.forName("akt").getMethod("a").invoke(null);
        Class<?> portal=Class.forName("dlr"),processor=Class.forName("btz"),pos=Class.forName("jd");
        Object origin=pos.getConstructor(int.class,int.class,int.class).newInstance(0,64,0);
        System.setOut(new PrintStream(new FileOutputStream(FileDescriptor.out),true));
        for(int delay:new int[]{0,1,80})for(int scenario=0;scenario<4;scenario++) {
            Object p=Proxy.newProxyInstance(portal.getClassLoader(),new Class<?>[]{portal},(proxy,method,values)->{
                if(method.getName().equals("a")&&values.length==2)return delay;
                if(method.isDefault())return InvocationHandler.invokeDefault(proxy,method,values);
                throw new UnsupportedOperationException(method.toString());
            });
            Object state=processor.getConstructor(portal,pos).newInstance(p,origin);
            for(int tick=0;tick<90;tick++) {
                boolean inside=scenario==0||scenario==1&&tick%9<6||scenario==2&&tick<5||scenario==3;
                boolean allowed=scenario!=3||tick>=5;
                processor.getMethod("a",boolean.class).invoke(state,inside);
                boolean triggered=(boolean)processor.getMethod("a",Class.forName("aqu"),Class.forName("bsr"),boolean.class).invoke(state,null,null,allowed);
                System.out.printf("portal_processor %d %d %d %s %s %s %s %s %s%n",delay,scenario,tick,inside,allowed,triggered,processor.getMethod("d").invoke(state),processor.getMethod("e").invoke(state),processor.getMethod("b").invoke(state));
            }
        }
        Class<?> ut=Class.forName("sun.misc.Unsafe");Field uf=ut.getDeclaredField("theUnsafe");uf.setAccessible(true);Object unsafe=uf.get(null);
        Object entity=ut.getMethod("allocateInstance",Class.class).invoke(unsafe,Class.forName("aqv"));
        Method cooldownTick=Class.forName("bsr").getDeclaredMethod("M");cooldownTick.setAccessible(true);
        Object p=Proxy.newProxyInstance(portal.getClassLoader(),new Class<?>[]{portal},(proxy,method,values)->null);
        for(int scenario=0;scenario<2;scenario++) {
            Class.forName("bsr").getMethod("ay").invoke(entity);
            for(int tick=0;tick<15;tick++) {
                boolean contact=scenario==1&&tick<7;
                if(contact)Class.forName("bsr").getMethod("a",portal,pos).invoke(entity,p,origin);
                cooldownTick.invoke(entity);
                System.out.printf("portal_player_cooldown %d %d %s %s%n",scenario,tick,contact,Class.forName("bsr").getMethod("az").invoke(entity));
            }
        }
        Object rules=Class.forName("dcs").getConstructor().newInstance();
        for(String name:new String[]{"E","F"})System.out.printf("portal_player_delay %s %s%n",name,Class.forName("dcs").getMethod("c",Class.forName("dcs$e")).invoke(rules,Class.forName("dcs").getField(name).get(null)));
    }
}
