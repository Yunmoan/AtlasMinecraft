/** Development-only shell; all chest menu operations remain original JAR code. */
import com.mojang.authlib.GameProfile;
public final class AtlasChestMenuPlayer extends cmx {
 public dcw level; public cmw inventory; public boolean creative;
 public AtlasChestMenuPlayer(dcw level) {super(level,new jd(0,0,0),0,new GameProfile(new java.util.UUID(0,2),"ChestProbe"));}
 @Override public boolean R_(){return false;}
 @Override public boolean f(){return creative;}
 @Override public boolean fL(){return creative;}
 @Override public dcw dO(){return level;}
 @Override public cmw fY(){return inventory;}
 @Override public void a(cuq carried,cuq slot,cqd button){}
}
