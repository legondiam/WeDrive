<template>
  <div class="landing-page min-h-screen bg-background text-foreground">
    <header class="fixed inset-x-0 top-0 z-50 border-b border-white/10 bg-[linear-gradient(180deg,#171717_0%,#141414_100%)] shadow-[0_10px_30px_rgba(0,0,0,0.18)]">
      <div class="flex w-full items-center justify-between px-6 py-4 md:px-8 xl:px-10">
        <RouterLink to="/" class="flex items-center gap-3">
          <BrandCloudIcon class="h-8 w-8 text-white" />
          <div class="flex flex-col">
            <span class="font-brand text-[22px] font-bold tracking-[-0.04em] text-white">WeDrive</span>
            <span class="text-[10px] font-medium uppercase tracking-[0.28em] text-white/40">
              Open Source Private Cloud Drive
            </span>
          </div>
        </RouterLink>
      </div>
    </header>

    <main class="mx-auto w-full max-w-[1480px] px-6 pb-12 pt-24 md:px-8 md:pb-16 md:pt-28 xl:px-12">
      <section class="poster-grid grid min-h-[calc(100vh-116px)] gap-8">
        <div class="poster-copy flex min-h-[420px] flex-col justify-between rounded-[34px] border border-zinc-200/70 bg-white/92 p-7 shadow-[0_24px_80px_rgba(15,23,42,0.06)] md:p-10">
          <div class="max-w-[920px]">
          <div class="flex flex-col gap-3">
            <div>
              <p class="font-brand text-[24px] font-bold tracking-[-0.05em] text-zinc-950 md:text-[30px]">
                WeDrive
              </p>
              <p class="mt-1 text-[11px] font-medium uppercase tracking-[0.34em] text-slate-500 md:text-[12px]">
                Open Source Private Cloud Drive
              </p>
            </div>
          </div>
          <h1 class="mt-5 font-brand text-[58px] font-bold leading-[0.92] tracking-[-0.07em] text-zinc-950 md:text-[84px] lg:max-w-[860px] lg:text-[118px] xl:text-[132px]">
            <TextGenerateEffect
              as="span"
              text="你的数据，"
              class="block"
            />
            <TextGenerateEffect
              as="span"
              text="只属于你。"
              class="mt-1 block"
              :start-delay="220"
            />
          </h1>
          <p class="mt-7 max-w-[580px] text-[15px] leading-[1.95] text-zinc-600 md:text-[17px]">
            WeDrive 是一个面向个人与小团队的私有云盘系统，用更轻的界面承接更快的上传、分享与归档体验。
          </p>
          </div>

          <div class="mt-10 flex flex-wrap items-center gap-3 isolate">
            <template v-if="!hasToken">
              <RouterLink
                to="/login"
                class="rounded-full bg-zinc-950 px-6 py-3 text-[14px] font-medium text-white transition-all duration-200 hover:bg-zinc-800"
              >
                前往登录
              </RouterLink>
              <RouterLink
                to="/register"
                class="rounded-full border border-zinc-200 bg-white px-6 py-3 text-[14px] font-medium text-zinc-700 transition-all duration-200 hover:border-zinc-300 hover:bg-zinc-50 hover:text-zinc-900"
              >
                创建账户
              </RouterLink>
            </template>
            <InteractiveHoverButton
              v-else
              :as="RouterLink"
              to="/drive"
              text="立即体验"
              class="border-zinc-300 bg-zinc-50 text-zinc-950 hover:border-zinc-400"
            />
          </div>

          <div class="mt-12 grid gap-4 border-t border-zinc-200/90 pt-6 md:grid-cols-3">
            <div>
              <p class="text-[12px] uppercase tracking-[0.24em] text-zinc-400">Performance</p>
              <p class="mt-2 text-[16px] font-medium text-zinc-900">并发稳定运行 / 响应保持轻快</p>
            </div>
            <div>
              <p class="text-[12px] uppercase tracking-[0.24em] text-zinc-400">Storage</p>
              <p class="mt-2 text-[16px] font-medium text-zinc-900">S3 兼容 / 对象存储</p>
            </div>
            <div>
              <p class="text-[12px] uppercase tracking-[0.24em] text-zinc-400">Sharing</p>
              <p class="mt-2 text-[16px] font-medium text-zinc-900">快速分享 / 秒传归档</p>
            </div>
          </div>
        </div>

        <div class="support-grid grid gap-8 lg:grid-cols-[minmax(0,1.08fr)_minmax(360px,0.92fr)]">
          <div class="hero-panel overflow-hidden rounded-[34px] border border-zinc-200/80 bg-white/88 p-6 shadow-[0_20px_60px_rgba(15,23,42,0.06)] backdrop-blur-sm md:p-7">
            <div class="flex items-center justify-between text-[12px] uppercase tracking-[0.26em] text-slate-500">
              <span>Go + Gin</span>
              <span>Runtime</span>
            </div>
            <div class="mt-6 flex items-center justify-between gap-6">
              <div>
                <h2 class="hero-title text-[28px] font-bold tracking-[-0.04em]">高并发访问</h2>
                <p class="mt-2 max-w-[360px] text-[14px] leading-[1.9] text-zinc-500">Go 并发，Gin 路由，响应保持轻快。</p>
              </div>
              <svg viewBox="0 0 2586 1024" aria-hidden="true" class="go-brand-mark">
                <path d="M192.026947 345.303579c-4.958316 0-6.144-2.479158-3.718736-6.197895l26.031157-33.414737c2.425263-3.718737 8.623158-6.197895 13.581474-6.197894h442.260211c4.958316 0 6.197895 3.772632 3.772631 7.437473l-21.018947 32.175158c-2.533053 3.772632-8.730947 7.437474-12.449684 7.437474l-448.458106-1.239579zM4.958316 459.290947c-4.958316 0-6.144-2.479158-3.664842-6.144l25.977263-33.468631c2.479158-3.772632 8.623158-6.197895 13.635368-6.197895H605.776842c4.904421 0 7.383579 3.772632 6.144 7.383579l-9.916631 29.749895c-1.239579 5.012211-6.197895 7.491368-11.156211 7.491368L4.958316 459.290947zM304.720842 573.278316c-4.904421 0-6.144-3.772632-3.664842-7.437474l17.300211-30.989474c2.479158-3.718737 7.491368-7.383579 12.395789-7.383579h247.808c4.958316 0 7.383579 3.664842 7.383579 8.623158l-2.425263 29.749895c0 4.958316-4.958316 8.623158-8.730948 8.623158l-270.012631-1.185684zM1590.649263 323.045053c-78.039579 19.833263-131.287579 34.654316-208.141474 54.487579-18.593684 4.958316-19.779368 6.197895-35.947789-12.39579-18.539789-21.018947-32.175158-34.708211-58.206316-47.050105-78.039579-38.373053-153.6-27.324632-224.202105 18.593684-84.237474 54.433684-127.568842 135.006316-126.383158 235.358316 1.293474 99.058526 69.416421 180.870737 167.235368 194.45221 84.291368 11.156211 154.893474-18.539789 210.620632-81.758315 11.156211-13.635368 21.018947-28.510316 33.468632-45.810527h-239.077053c-26.031158 0-32.175158-16.114526-23.552-37.133473 16.060632-38.480842 45.810526-102.831158 63.164632-135.060211a33.414737 33.414737 0 0 1 30.989473-19.833263h450.937263c-2.479158 33.468632-2.479158 66.883368-7.437473 100.352-13.635368 89.195789-47.104 170.954105-101.645474 242.795789-89.141895 117.706105-205.608421 190.787368-353.010526 210.620632-121.424842 16.060632-234.172632-7.437474-333.285053-81.758316-91.621053-69.416421-143.683368-161.091368-157.264842-275.024842-16.168421-135.060211 23.498105-256.377263 105.256421-362.981053 87.956211-115.226947 204.422737-188.254316 346.866526-214.339368 116.466526-21.018947 227.974737-7.383579 328.326737 60.739368 65.643789 43.331368 112.747789 102.831158 143.683369 174.618948 7.437474 11.210105 2.425263 17.354105-12.39579 21.126737" fill="#00ACD7" />
                <path d="M2295.538526 507.634526c-1.239579-16.114526-1.239579-28.564211-3.664842-40.906105-22.312421-122.610526-135.114105-191.973053-252.766316-164.75621-115.226947 25.977263-189.547789 99.166316-216.764631 215.578947-22.366316 96.633263 24.737684 194.506105 113.933474 234.172631 68.122947 29.696 136.245895 25.977263 201.889684-7.491368 97.926737-50.822737 151.228632-130.048 157.372631-236.597895m-294.80421 500.466527c-112.747789-2.479158-215.578947-34.654316-302.349474-108.975158C1625.357474 835.907368 1579.439158 755.388632 1564.618105 659.994947 1542.305684 520.030316 1580.732632 396.126316 1665.024 285.911579c90.327579-118.945684 199.410526-180.870737 346.812632-206.95579C2138.165895 56.697263 2257.111579 69.093053 2364.901053 142.228211c97.872842 66.937263 158.558316 157.372632 174.726736 276.264421 21.018947 167.235368-27.324632 303.481263-142.497684 419.947789-81.758316 82.997895-182.056421 135.060211-297.337263 158.558316-33.414737 6.197895-66.829474 7.491368-99.058526 11.15621" fill="#00ACD7" />
              </svg>
            </div>
            <div class="code-visual mt-7" v-html="goCodeHtml" />
            <p class="mt-6 max-w-[420px] text-[14px] leading-[1.9] text-zinc-500">多人同时上传、下载、分享时，依然保持响应稳定。面向高频访问场景的轻量服务层。</p>
          </div>

          <div class="feature-stack grid gap-8">
            <article class="feature-panel flex h-full min-h-[360px] flex-col rounded-[34px] border border-zinc-200/80 bg-white/92 p-6 shadow-[0_18px_50px_rgba(15,23,42,0.05)] md:p-8">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-[12px] font-medium uppercase tracking-[0.28em] text-slate-500">Storage</p>
                  <h2 class="mt-3 text-[28px] font-bold tracking-[-0.04em] text-zinc-900">OSS 对象存储</h2>
                </div>
                <span class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-[12px] text-slate-600">Object Storage</span>
              </div>
              <div class="storage-visual mt-8">
                <div class="storage-mark">
                  <svg viewBox="0 0 2048 1024" aria-hidden="true" class="storage-mark__word">
                    <path d="M1358.56128 512c34.57024 0.90112 67.7888 6.5536 98.38592 22.79424 30.0032 15.95392 51.44576 39.19872 56.99584 74.11712 6.0416 37.9904-5.20192 70.10304-34.93888 94.69952-25.76384 21.2992-56.58624 30.47424-89.35424 33.9968-26.91072 2.88768-53.73952 2.6624-80.2816-3.01056-37.35552-7.96672-70.51264-22.97856-92.07808-56.64768-22.6304-35.328-18.16576-85.52448 10.0352-116.67456 22.03648-24.3712 50.44224-36.78208 81.75616-43.04896 16.26112-3.25632 32.9728-4.21888 49.47968-6.22592z m-297.24672 4.54656c6.63552-0.12288 9.1136 1.61792 9.0112 8.76544-0.36864 30.94528-0.14336 61.91104-0.14336 92.85632-0.04096 31.21152 0.06144 62.42304-0.1024 93.63456-0.06144 13.23008-7.55712 22.3232-19.49696 23.12192a26.64448 26.64448 0 0 1-13.45536-3.11296c-61.46048-32.39936-122.79808-64.98304-184.1152-97.62816-14.70464-7.82336-17.38752-6.26688-17.36704 10.48576 0.04096 27.21792-0.12288 54.43584 0.08192 81.65376 0.06144 6.36928-1.26976 9.29792-8.704 9.09312-19.456-0.53248-38.93248-0.24576-58.40896 0.04096-6.18496 0.1024-8.8064-1.6384-8.8064-8.3968 0.2048-61.66528 0.02048-123.33056 0.2048-185.01632 0.02048-14.04928 5.79584-22.07744 17.53088-24.1664 4.5056-0.79872 10.1376 0.04096 14.17216 2.17088 55.52128 29.20448 110.85824 58.7776 166.23616 88.24832 7.92576 4.1984 15.83104 8.3968 23.8592 12.41088 8.11008 4.03456 11.5712 1.9456 11.63264-7.31136 0.18432-28.4672 0.14336-56.91392 0.2048-85.38112 0-11.22304 0.02048-11.264 11.48928-11.32544l37.45792 0.04096 18.71872-0.18432z m93.53216 0.08192c5.30432-0.24576 6.81984 1.98656 6.79936 7.0656-0.1024 68.4032-0.1024 136.8064 0 205.18912 0.02048 4.9152-1.67936 6.67648-6.4512 6.5536-9.216-0.16384-18.432-0.04096-28.99968-0.04096V526.336c0-9.60512 0-9.6256 9.9328-9.64608 6.2464-0.02048 12.4928 0.24576 18.71872-0.04096z m-516.01408 0c19.21024 0.47104 38.42048 0.36864 57.6512 0.06144 5.9392-0.1024 7.86432 1.98656 7.84384 7.7824a48400.384 48400.384 0 0 0 0.06144 202.9568c0 5.85728-1.96608 7.96672-7.72096 7.90528-20.21376-0.2048-40.448-0.28672-60.66176 0-5.89824 0.06144-6.9632-2.4576-6.9632-7.65952 0.18432-33.93536 0.1024-67.8912 0.1024-101.84704 0-33.19808 0.24576-66.41664-0.18432-99.61472-0.1024-7.7824 2.47808-9.74848 9.87136-9.56416z m-100.39296 4.79232c17.77664-11.61216 35.0208 2.08896 34.85696 19.2512-0.55296 62.42304-0.02048 124.84608-0.14336 187.24864 0 2.43712-2.56 6.92224-3.9936 6.9632-11.5712 0.2048-23.1424 0.28672-34.87744 0.3072H498.4832v-61.82912c0-9.97376 0.08192-19.968-0.08192-29.94176-0.14336-8.6016-4.62848-11.264-11.89888-7.02464-27.136 15.79008-54.39488 31.41632-81.22368 47.7184-13.0048 7.8848-24.59648 7.7824-37.4784-0.45056-24.9856-15.9744-50.4832-31.08864-75.83744-46.44864-8.2944-5.0176-12.96384-2.17088-12.98432 7.82336-0.08192 27.70944 0 55.41888-0.02048 83.1488 0 1.9456-0.2048 3.8912-0.36864 6.43072h-73.35936c-0.12288-2.82624-0.34816-5.46816-0.34816-8.13056 0-61.41952 0.08192-122.83904-0.1024-184.25856-0.02048-9.40032 1.8432-17.46944 10.58816-22.24128 8.82688-4.83328 16.95744-2.62144 25.25184 2.47808 47.34976 29.0816 94.9248 57.79456 142.2336 86.95808 5.0176 3.072 8.58112 2.17088 12.84096-0.45056 37.80608-23.18336 75.63264-46.32576 113.4592-69.48864 9.76896-5.98016 19.6608-11.79648 29.2864-18.06336z m832.14336 22.69184c-44.60544-0.04096-75.5712 5.79584-102.7072 25.84576-34.95936 25.82528-36.49536 83.12832-2.80576 110.40768 20.25472 16.384 43.64288 24.2688 69.2224 26.76736 33.13664 3.21536 65.6384 1.57696 96.54272-12.36992 14.60224-6.59456 27.38176-15.9744 35.86048-29.61408 21.83168-35.10272 11.776-77.27104-23.06048-99.86048-25.88672-16.7936-54.84544-21.11488-73.05216-21.17632z" fill="#C72C48" />
                  </svg>
                  <svg viewBox="0 0 1024 1024" aria-hidden="true" class="storage-mark__glyph">
                    <path d="M638.855218 56.525807s99.268136 159.838523 132.357514 216.623262a2.523766 2.523766 0 0 1 0 2.944394 2.383557 2.383557 0 0 1-3.50523 0L596.231612 97.326693l42.623606-40.800886z" fill="#dd113c" />
                    <path d="M346.518971 639.655999a588.878771 588.878771 0 0 1 116.654081-165.446893 597.291325 597.291325 0 0 1 58.32704-51.176369v126.188308L346.518971 639.655999zM245.568325 756.590498l275.931767-140.209231v321.079139l62.112689 80.760517v-434.648616l37.716283-19.489084a187.179324 187.179324 0 0 0 51.456788-296.121896L530.753901 119.479752a31.547077 31.547077 0 0 1 1.542302-44.446327 31.687286 31.687286 0 0 1 44.586535 1.542302l19.909711 20.750966 42.062769-40.941095c-50.335114-65.337502-112.167385-57.065157-147.64032-24.396407a90.575163 90.575163 0 0 0-3.925859 127.870819l143.574253 149.60325a128.151237 128.151237 0 0 1-28.041846 197.414597l-19.489083 10.095065V314.090164A649.589368 649.589368 0 0 0 245.568325 755.889452v0.701046z" fill="#dd113c" />
                    <path d="M583.612781 583.432097v65.617921l-62.112689 31.547077v-65.197293z" fill="#dd113c" />
                  </svg>
                </div>
                <p class="storage-copy text-[14px] leading-[1.9] text-zinc-500">
                  S3 协议兼容。由 MinIO 驱动的工业级存储。适合承接文件上传、归档与长期管理。
                </p>
              </div>
            </article>

            <article class="feature-panel flex h-full min-h-[360px] flex-col rounded-[34px] border border-zinc-200/80 bg-white/92 p-6 shadow-[0_18px_50px_rgba(15,23,42,0.05)] md:p-8">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-[12px] font-medium uppercase tracking-[0.28em] text-slate-500">Upload</p>
                  <h2 class="mt-3 text-[28px] font-bold tracking-[-0.04em] text-zinc-900">秒传与快速入库</h2>
                </div>
                <span class="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-[12px] text-slate-600">Instant Upload</span>
              </div>
              <div class="mt-8 rounded-[24px] border border-zinc-200 bg-zinc-50/80 p-5">
                <div class="flex items-center justify-between text-[12px] uppercase tracking-[0.24em] text-zinc-500">
                  <span>Hash Match</span>
                  <span class="instant-chip">Hit</span>
                </div>
                <div class="mt-5 overflow-hidden rounded-full bg-zinc-200">
                  <div class="instant-bar h-3 rounded-full bg-[linear-gradient(90deg,#dc2626_0%,#ef4444_36%,#f87171_68%,#fee2e2_100%)]" />
                </div>
              </div>
              <p class="mt-4 max-w-[520px] text-[14px] leading-[1.9] text-zinc-500">
                通过文件指纹命中缓存，减少重复传输，让重复上传更快完成确认与写入。
              </p>
              <p class="mt-3 max-w-[420px] text-[14px] leading-[1.9] text-zinc-500">
                把重复文件处理成本降到更低。
              </p>
            </article>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { RouterLink } from 'vue-router'
import { computed } from 'vue'
import BrandCloudIcon from '@/components/icons/BrandCloudIcon.vue'
import InteractiveHoverButton from '@/components/ui/button/InteractiveHoverButton.vue'
import TextGenerateEffect from '@/components/ui/text-generate-effect/TextGenerateEffect.vue'

const hasToken = computed(() => Boolean(localStorage.getItem('accessToken')))
const goCodeHtml = `<pre class="shiki dark-plus" style="background-color:#0B0E14;color:#D4D4D4" tabindex="0"><code><span class="line"><span style="color:#9CDCFE">publicGroup</span><span style="color:#D4D4D4"> := </span><span style="color:#9CDCFE">r</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">Group</span><span style="color:#D4D4D4">(</span><span style="color:#CE9178">"/api/v1"</span><span style="color:#D4D4D4">)</span></span>
<span class="line"><span style="color:#9CDCFE">publicGroup</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">Use</span><span style="color:#D4D4D4">(</span><span style="color:#9CDCFE">middleware</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">TimeoutMiddleware</span><span style="color:#D4D4D4">(</span><span style="color:#B5CEA8">3</span><span style="color:#D4D4D4"> * </span><span style="color:#9CDCFE">time</span><span style="color:#D4D4D4">.</span><span style="color:#9CDCFE">Second</span><span style="color:#D4D4D4">))</span></span>
<span class="line"></span>
<span class="line"><span style="color:#9CDCFE">privateGroup</span><span style="color:#D4D4D4"> := </span><span style="color:#9CDCFE">publicGroup</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">Group</span><span style="color:#D4D4D4">(</span><span style="color:#CE9178">"/"</span><span style="color:#D4D4D4">)</span></span>
<span class="line"><span style="color:#9CDCFE">privateGroup</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">Use</span><span style="color:#D4D4D4">(</span><span style="color:#9CDCFE">middleware</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">AuthMiddleware</span><span style="color:#D4D4D4">())</span></span>
<span class="line"><span style="color:#D4D4D4">{</span></span>
<span class="line"><span style="color:#9CDCFE">    privateGroup</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">GET</span><span style="color:#D4D4D4">(</span><span style="color:#CE9178">"/file/list"</span><span style="color:#D4D4D4">, </span><span style="color:#9CDCFE">fileHandler</span><span style="color:#D4D4D4">.</span><span style="color:#9CDCFE">GetUserFile</span><span style="color:#D4D4D4">)</span></span>
<span class="line"><span style="color:#9CDCFE">    privateGroup</span><span style="color:#D4D4D4">.</span><span style="color:#DCDCAA">GET</span><span style="color:#D4D4D4">(</span><span style="color:#CE9178">"/file/download/:ID"</span><span style="color:#D4D4D4">, </span><span style="color:#9CDCFE">fileHandler</span><span style="color:#D4D4D4">.</span><span style="color:#9CDCFE">GetDownloadURL</span><span style="color:#D4D4D4">)</span></span>
<span class="line"><span style="color:#D4D4D4">}</span></span></code></pre>`
</script>

<style scoped>
.landing-page {
  background:
    radial-gradient(circle at 8% 10%, rgba(148, 163, 184, 0.16), transparent 28%),
    radial-gradient(circle at 92% 18%, rgba(226, 232, 240, 0.88), transparent 26%),
    linear-gradient(135deg, #fcfcfd 0%, #f6f7f9 52%, #f2f4f7 100%);
}

.landing-page header,
.landing-page h1,
.landing-page h2,
.landing-page .storage-mark,
.landing-page .go-brand-mark,
.landing-page .instant-bar,
.landing-page .instant-chip,
.landing-page .code-visual {
  user-select: none;
}

.code-visual,
.code-visual * {
  user-select: text;
}

.poster-grid {
  align-items: stretch;
}

.hero-panel,
.feature-panel {
  position: relative;
}

.hero-panel::after,
.feature-panel::after {
  content: '';
  position: absolute;
  inset: 1px;
  pointer-events: none;
  border-radius: inherit;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.28), rgba(255, 255, 255, 0));
}

.hero-panel::after {
  display: none;
}

.code-visual {
  overflow: hidden;
  border-radius: 14px;
  background: #0B0E14;
  border: 1px solid #0B0E14;
  box-shadow: none;
}

:deep(.code-visual pre) {
  margin: 0;
  padding: 1.2rem 1.3rem;
  background: transparent !important;
  overflow: hidden;
}

:deep(.code-visual code) {
  font-size: 13px;
  line-height: 1.75;
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
}

:deep(.code-visual .line) {
  display: block;
}

.go-brand-mark {
  width: min(11rem, 36%);
  height: auto;
  flex: none;
}

.hero-title {
  color: rgb(82 82 91);
}

.storage-visual {
  display: flex;
  min-height: 220px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(248, 250, 252, 0.96), rgba(241, 245, 249, 0.8)),
    radial-gradient(circle at top, rgba(148, 163, 184, 0.14), transparent 40%);
  padding: 1.5rem;
}

.storage-mark {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.storage-mark__glyph {
  width: 5.6rem;
  height: 5.6rem;
  flex: none;
}

.storage-mark__word {
  width: min(20rem, 100%);
  height: auto;
}

.storage-copy {
  margin-top: 1.25rem;
  max-width: 22rem;
  text-align: center;
}

.instant-chip {
  border-radius: 999px;
  background: rgba(100, 116, 139, 0.12);
  color: rgb(71 85 105);
  padding: 4px 10px;
}

.instant-bar {
  width: 0;
  animation: instantFill 2.6s cubic-bezier(0.16, 1, 0.3, 1) infinite;
}

@keyframes instantFill {
  0% {
    width: 0;
    opacity: 0.7;
  }
  16% {
    width: 100%;
    opacity: 1;
  }
  82% {
    width: 100%;
    opacity: 1;
  }
  100% {
    width: 0;
    opacity: 0;
  }
}

@media (max-width: 1024px) {
  .poster-grid {
    min-height: auto;
  }
}

@media (max-width: 900px) {
  .storage-visual {
    min-height: 160px;
  }
}
</style>
