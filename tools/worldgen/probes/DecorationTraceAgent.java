package atlasprobe;

import java.io.*;
import java.lang.instrument.*;
import java.security.ProtectionDomain;
import org.objectweb.asm.*;

/** Observe original applyBiomeDecoration entry/exit; does not replace its algorithm. */
public final class DecorationTraceAgent {
    public static void premain(String path, Instrumentation instrumentation) throws Exception {
        DecorationTraceSink.open(path);
        instrumentation.addTransformer(new ClassFileTransformer() {
            @Override public byte[] transform(ClassLoader loader, String name, Class<?> redefined,
                                              ProtectionDomain domain, byte[] bytes) {
                if (!name.equals("duz") && !name.equals("eiv") && !name.equals("ebq") && !name.equals("dvi") && !name.equals("dkr") && !name.equals("aqb") && !name.equals("aqu")) return null;
                boolean chunkMapClass=name.equals("aqb"), levelClass=name.equals("aqu");
                boolean magmaClass=name.equals("dkr");
                boolean chunkClass = name.equals("dvi");
                boolean featureClass = name.equals("eiv");
                boolean configuredClass = name.equals("ebq");
                ClassReader reader = new ClassReader(bytes);
                ClassWriter writer = new ClassWriter(reader, ClassWriter.COMPUTE_MAXS);
                reader.accept(new ClassVisitor(Opcodes.ASM8, writer) {
                    @Override public MethodVisitor visitMethod(int access, String name, String descriptor,
                                                               String signature, String[] exceptions) {
                        MethodVisitor original = super.visitMethod(access, name, descriptor, signature, exceptions);
                        if(chunkMapClass&&name.equals("a")&&descriptor.equals("(Lduy;)Z"))return new MethodVisitor(Opcodes.ASM8,original){
                            @Override public void visitCode(){super.visitCode();super.visitVarInsn(Opcodes.ALOAD,0);super.visitVarInsn(Opcodes.ALOAD,1);super.visitMethodInsn(Opcodes.INVOKESTATIC,"atlasprobe/DecorationTraceSink","beforeChunkSave","(Ljava/lang/Object;Ljava/lang/Object;)V",false);}
                        };
                        if(levelClass&&(name.equals("d")&&descriptor.equals("(Ljd;Ldfy;)V")||name.equals("a")&&descriptor.equals("(Ljd;Lepd;)V")))return new MethodVisitor(Opcodes.ASM8,original){
                            @Override public void visitCode(){super.visitCode();super.visitLdcInsn(name.equals("d")?"block":"fluid");for(int i=0;i<3;i++)super.visitVarInsn(Opcodes.ALOAD,i);super.visitMethodInsn(Opcodes.INVOKESTATIC,"atlasprobe/DecorationTraceSink","scheduledDispatch","(Ljava/lang/String;Ljava/lang/Object;Ljava/lang/Object;Ljava/lang/Object;)V",false);}
                        };
                        if(magmaClass&&name.equals("a")&&descriptor.equals("(Ldtc;Lji;Ldtc;Ldcx;Ljd;Ljd;)Ldtc;"))return new MethodVisitor(Opcodes.ASM8,original){
                            @Override public void visitCode(){super.visitCode();for(int i:new int[]{5,2,3,4})super.visitVarInsn(Opcodes.ALOAD,i);super.visitMethodInsn(Opcodes.INVOKESTATIC,"atlasprobe/DecorationTraceSink","magma","(Ljava/lang/Object;Ljava/lang/Object;Ljava/lang/Object;Ljava/lang/Object;)V",false);}
                        };
                        if (chunkClass && (name.equals("H") && descriptor.equals("()V") || name.equals("c") && descriptor.equals("(J)V"))) {
                            String phase=name.equals("H")?"postprocess":"unpack";
                            return new MethodVisitor(Opcodes.ASM8,original) {
                                private void event(String suffix){
                                    super.visitLdcInsn(phase+suffix);super.visitVarInsn(Opcodes.ALOAD,0);
                                    super.visitMethodInsn(Opcodes.INVOKESTATIC,"atlasprobe/DecorationTraceSink","activation","(Ljava/lang/String;Ljava/lang/Object;)V",false);
                                }
                                @Override public void visitCode(){super.visitCode();event("_begin");}
                                @Override public void visitInsn(int opcode){if(opcode==Opcodes.RETURN)event("_end");super.visitInsn(opcode);}
                            };
                        }
                        if (configuredClass && name.equals("a") && descriptor.equals("(Ldds;Lduz;Layw;Ljd;)Z")) {
                            return new MethodVisitor(Opcodes.ASM8, original) {
                                @Override public void visitCode() {
                                    super.visitCode();
                                    super.visitVarInsn(Opcodes.ALOAD, 4);
                                    super.visitMethodInsn(Opcodes.INVOKESTATIC,"atlasprobe/DecorationTraceSink","configuredOrigin","(Ljava/lang/Object;)V",false);
                                }
                            };
                        }
                        if (featureClass && name.equals("b") && descriptor.equals("(Ldds;Lduz;Layw;Ljd;)Z")) {
                            return new MethodVisitor(Opcodes.ASM8, original) {
                                @Override public void visitCode() {
                                    super.visitCode();
                                    super.visitVarInsn(Opcodes.ALOAD, 0);
                                    super.visitVarInsn(Opcodes.ALOAD, 2);
                                    super.visitVarInsn(Opcodes.ALOAD, 3);
                                    super.visitVarInsn(Opcodes.ALOAD, 1);
                                    super.visitMethodInsn(Opcodes.INVOKESTATIC, "atlasprobe/DecorationTraceSink", "featureBegin",
                                                          "(Ljava/lang/Object;Ljava/lang/Object;Ljava/lang/Object;Ljava/lang/Object;)V", false);
                                }
                                @Override public void visitInsn(int opcode) {
                                    if (opcode == Opcodes.IRETURN) super.visitMethodInsn(Opcodes.INVOKESTATIC,
                                        "atlasprobe/DecorationTraceSink", "featureEnd", "()V", false);
                                    super.visitInsn(opcode);
                                }
                            };
                        }
                        if (!name.equals("a") || !descriptor.equals("(Ldds;Lduy;Lddq;)V")) return original;
                        return new MethodVisitor(Opcodes.ASM8, original) {
                            private void event(String phase) {
                                super.visitLdcInsn(phase);
                                if (phase.equals("begin")) super.visitVarInsn(Opcodes.ALOAD, 2);
                                else super.visitInsn(Opcodes.ACONST_NULL);
                                super.visitMethodInsn(Opcodes.INVOKESTATIC, "atlasprobe/DecorationTraceSink", "event",
                                                      "(Ljava/lang/String;Ljava/lang/Object;)V", false);
                            }
                            @Override public void visitCode() { super.visitCode(); event("begin"); }
                            @Override public void visitInsn(int opcode) {
                                if (opcode == Opcodes.RETURN) event("end");
                                super.visitInsn(opcode);
                            }
                        };
                    }
                }, 0);
                DecorationTraceSink.note("instrumented " + name);
                return writer.toByteArray();
            }
        });
    }

}
